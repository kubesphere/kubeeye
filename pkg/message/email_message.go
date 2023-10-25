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
	"net/smtp"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	var secret corev1.Secret
	err := e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: constant.DefaultNamespace,
		Name:      e.SecretKey,
	}, &secret)

	if err != nil {
		klog.Error("failed to get secret", err)
		return
	}

	auth := smtp.PlainAuth("", string(secret.Data["username"]), string(secret.Data["password"]), e.Address)
	err = e.SendMsg(auth, event)
	if err != nil {
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

func (e *EmailMessageHandler) SendMsg(a smtp.Auth, eve *conf.MessageEvent) error {
	dial, err := smtp.Dial(fmt.Sprintf("%s:%d", e.Address, e.Port))
	if err != nil {
		return err
	}
	defer dial.Close()
	err = dial.Hello(e.Address)
	if err != nil {
		return err
	}
	if ok, _ := dial.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: e.Address}
		if err = dial.StartTLS(config); err != nil {
			return err
		}
	}

	if a != nil {
		auth, _ := dial.Extension("AUTH")
		if !auth {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = dial.Auth(a); err != nil {
			return err
		}
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
