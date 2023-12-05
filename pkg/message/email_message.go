package message

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"mime"
	"net"
	"net/smtp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type EmailMessageHandler struct {
	// 可以添加处理器需要的属性
	*conf.EmailConfig
	client.Client
}

func NewEmailMessageOptions(event *conf.EmailConfig, c client.Client) *EmailMessageHandler {
	return &EmailMessageHandler{
		EmailConfig: event,
		Client:      c,
	}
}

func (e *EmailMessageHandler) HandleMessageEvent(event *conf.MessageEvent) {

	if err := e.Vail(); err != nil {
		klog.Error("failed to vail params", err)
		return
	}

	if err := e.SendMsg(event); err != nil {
		klog.Error("failed to send email, err: ", err)
		return
	}
	klog.Info("send email success")
}

func (e *EmailMessageHandler) setMsg(me *conf.MessageEvent, to string) []byte {
	buffer := &bytes.Buffer{}
	_, _ = fmt.Fprintf(buffer, "From: %s\r\n", mime.QEncoding.Encode("utf-8", e.Fo))
	_, _ = fmt.Fprintf(buffer, "To: %s\r\n", mime.QEncoding.Encode("utf-8", to))
	_, _ = fmt.Fprintf(buffer, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", me.Title))
	_, _ = fmt.Fprintf(buffer, "Message-Id: %s\r\n", fmt.Sprintf("<%d.@%s>", time.Now().UnixNano(), e.Address))
	_, _ = fmt.Fprintf(buffer, "Date: %s\r\n", me.Timestamp.Format(time.RFC1123Z))
	_, _ = fmt.Fprintf(buffer, "Content-Type: text/html; charset=UTF-8;")
	_, _ = fmt.Fprintf(buffer, "MIME-Version: 1.0\r\n\r\n")
	_, _ = fmt.Fprintf(buffer, "%s", me.Content)
	return buffer.Bytes()
}

func (e *EmailMessageHandler) SendMsg(eve *conf.MessageEvent) error {

	var conn net.Conn
	var err error

	if e.Port == 465 {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", e.Address, e.Port), tlsConfig)
	} else {
		d := net.Dialer{}
		conn, err = d.Dial("tcp", fmt.Sprintf("%s:%d", e.Address, e.Port))
	}

	if err != nil {
		return err
	}
	dial, err := smtp.NewClient(conn, e.Address)
	if err != nil {
		return err
	}
	defer dial.Close()
	err = dial.Hello(e.Address)
	if err != nil {
		return err
	}
	//if ok, _ := dial.Extension("STARTTLS"); ok {
	//	config := &tls.Config{ServerName: e.Address}
	//	if err = dial.StartTLS(config); err != nil {
	//		return err
	//	}
	//}

	ok, param := dial.Extension("AUTH")
	if !ok {
		return errors.New("smtp: server doesn't support AUTH")
	}
	auth, err := e.auth(param)
	if err != nil {
		return err
	}

	if err = dial.Auth(auth); err != nil {
		return err
	}

	if err = dial.Mail(e.Fo); err != nil {
		return err
	}
	for _, addr := range e.To {
		if err = dial.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := dial.Data()
	if err != nil {
		return err
	}
	for _, s := range e.To {
		_, err = w.Write(e.setMsg(eve, s))
		if err != nil {
			klog.Errorf("unable send mail to %s", s)
		}
	}

	err = w.Close()
	if err != nil {
		return err
	}
	return dial.Quit()
}

func (e *EmailMessageHandler) Vail() error {
	if utils.IsEmptyValue(e.Address) {
		return errors.New("address is empty")
	}
	if e.Port == 0 {
		return errors.New("port  error")
	}
	if utils.IsEmptyValue(e.Fo) {
		return errors.New("fo is empty")
	}
	if utils.IsEmptyValue(e.To) {
		return errors.New("to is empty")
	}
	if utils.IsEmptyValue(e.SecretKey) {
		return errors.New("secretKey is empty")
	}

	return nil
}

type mailAuth struct {
	username, password string
}

func (e *EmailMessageHandler) auth(authType string) (smtp.Auth, error) {
	var secret corev1.Secret
	err := e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: constant.DefaultNamespace,
		Name:      e.SecretKey,
	}, &secret)
	if err != nil {
		return nil, err
	}

	for _, t := range strings.Split(authType, " ") {
		switch t {
		case "PLAIN":
			return smtp.PlainAuth("", string(secret.Data["username"]), string(secret.Data["password"]), e.Address), nil
			//case "LOGIN":
			//	return MailAuth(string(secret.Data["username"]), string(secret.Data["password"])), nil
		}
	}

	return nil, fmt.Errorf("unknown auth mechanism: %s", authType)
}

func MailAuth(username, password string) smtp.Auth {
	return &mailAuth{username, password}
}

func (a *mailAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte(a.username + "\x00" + a.password)
	return "LOGIN", resp, nil
}

func (a *mailAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challenge")
		}
	}
	return nil, nil
}
