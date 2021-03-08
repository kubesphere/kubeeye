package validator

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	ds "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"text/template"
	"time"
)

var ntpBox = (*packr.Box)(nil)

type NtpImageName struct {
	NtpImage string
}

func CheckNtp(ctx context.Context, ntpImage string) error {
	var tplWriter bytes.Buffer
	imageName := NtpImageName{NtpImage: ntpImage}

	dsTmplString, err := getNtpBox().FindString("ntp.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get ntp.yaml")
	}

	dsTemplate, err := template.New("ntp").Parse(dsTmplString)
	if dsTemplate == nil || err != nil {
		return errors.Wrap(err, "Failed to get ntp.yaml template")
	}
	err = dsTemplate.Execute(&tplWriter, imageName)
	if err != nil {
		return errors.Wrap(err, "Failed to render ntp.yaml template")
	}

	tplWriter.Bytes()

	pod := podParse(tplWriter.Bytes())

	//Create ntp
	_, err5 := createNtpDsSet(ctx, pod)
	if err5 != nil {
		return errors.Wrap(err5, "Failed to create ntp")
	}

	return nil
}
func getNtpBox() *packr.Box {
	if ntpBox == (*packr.Box)(nil) {
		ntpBox = packr.New("Ntp", "../../examples")
	}
	return ntpBox
}

func createNtpDsSet(ctx context.Context, conf *ds.DaemonSet) (*ds.DaemonSet, error) {
	kubeConf, configError := config.GetConfig()

	if configError != nil {
		logrus.Errorf("Error fetching KubeConfig: %v", configError)
		return nil, configError
	}

	api, err1 := kubernetes.NewForConfig(kubeConf)
	if err1 != nil {
		logrus.Errorf("Error fetching api: %v", err1)
		return nil, err1
	}
	listOpts := metav1.CreateOptions{}
	getOpts := metav1.GetOptions{}
	deleteOpts := metav1.DeleteOptions{}
	_, err2 := api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		fmt.Println("Installing Ntp ...")
		_, err3 := api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Create(ctx, conf, listOpts)
		if err3 != nil {
			return nil, err3
		}
		fmt.Println("Ntp Installation is successful. ")
		for i := 10; i > 0; i-- {
			status, err4 := api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Get(ctx, conf.ObjectMeta.Name, getOpts)
			if err4 != nil {
				return nil, err4
			}
			time.Sleep(time.Second * 5)

			if status.Status.DesiredNumberScheduled == status.Status.NumberReady {
				var podNames []string
				var nodeNames []string
				output, _ := exec.Command("/bin/sh", "-c", fmt.Sprintf("/usr/local/bin/kubectl --no-headers=true get pod -o wide | grep ntp | awk '{print $7}'")).CombinedOutput()
				nodeNames = strings.Split(string(output), "\n")
				//nodeNames = append(nodeNames,string(output))
				output1, _ := exec.Command("/bin/sh", "-c", fmt.Sprintf("/usr/local/bin/kubectl get pod | grep ntp | awk '{print $1}'")).CombinedOutput()
				podNames = strings.Split(string(output1), "\n")
				for i := 0; i < len(podNames)-1; i++ {
					output2, _ := exec.Command("/bin/sh", "-c", fmt.Sprintf("/usr/local/bin/kubectl logs %s", podNames[i])).CombinedOutput()
					fmt.Println(fmt.Sprintf("%s:", strings.TrimRight(string(nodeNames[i]), "\n")))
					fmt.Println(string(output2))
					time.Sleep(time.Millisecond * 500)
				}

				api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Delete(ctx, conf.ObjectMeta.Name, deleteOpts)
				break
			}
		}
		return nil, err3
	} else {
		fmt.Println("Please delete NTP service and try again.")
	}
	return nil, nil
}

func podParse(rawBytes []byte) *ds.DaemonSet {
	reader := bytes.NewReader(rawBytes)
	var conf *ds.DaemonSet
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&conf); err != nil {
			if err == io.EOF {
				break
			}
			return conf
		}
	}
	return conf
}
