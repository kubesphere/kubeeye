// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"bytes"
	"context"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	ds "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var configBox = (*packr.Box)(nil)

func Add(ctx context.Context) error {
	var rawBytes []byte

	// configMap create
	rawBytes, err := getConfigBox().Find("npd-rule.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get npd-rule.yaml")
	}
	config := Parse(rawBytes)
	_, err1 := createConfigMap(ctx, config)
	if err1 != nil {
		return errors.Wrap(err1, "Failed to create configmap")
	}

	// serviceAccount create
	saBytes, err := getConfigBox().Find("serviceAccount.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get serverAccount.yaml")
	}
	sa := saParse(saBytes)
	_, err2 := createServiceAccount(ctx, sa)
	if err2 != nil {
		return errors.Wrap(err2, "Failed to create serviceAccount")
	}

	// clusterRole create
	crBytes, err := getConfigBox().Find("clusterRole.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get clusterRole.yaml")
	}
	cr := crParse(crBytes)
	_, err3 := createClusterRole(ctx, cr)
	if err3 != nil {
		return errors.Wrap(err3, "Failed to create clusterRole")
	}

	// clusterRoleBinding create
	crbBytes, err := getConfigBox().Find("clusterRoleBinding.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get clusterRoleBinding.yaml")
	}
	crb := crbParse(crbBytes)
	_, err4 := createClusterRoleBinding(ctx, crb)
	if err4 != nil {
		return errors.Wrap(err4, "Failed to create clusterRoleBinding")
	}
	// daemonSet create
	dsBytes, err := getConfigBox().Find("daemonSet.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to get daemonSet.yaml")
	}
	ds := dsParse(dsBytes)
	_, err5 := createDaemonSet(ctx, ds)
	if err5 != nil {
		return errors.Wrap(err5, "Failed to create daemonSet")
	}
	return nil
}
func getConfigBox() *packr.Box {
	if configBox == (*packr.Box)(nil) {
		configBox = packr.New("Npd", "../../examples")
	}
	return configBox
}
func Parse(rawBytes []byte) *v1.ConfigMap {
	reader := bytes.NewReader(rawBytes)
	var conf *v1.ConfigMap
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
func saParse(rawBytes []byte) *v1.ServiceAccount {
	reader := bytes.NewReader(rawBytes)
	var conf *v1.ServiceAccount
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
func crParse(rawBytes []byte) *rbac.ClusterRole {
	reader := bytes.NewReader(rawBytes)
	var conf *rbac.ClusterRole
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
func crbParse(rawBytes []byte) *rbac.ClusterRoleBinding {
	reader := bytes.NewReader(rawBytes)
	var conf *rbac.ClusterRoleBinding
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
func dsParse(rawBytes []byte) *ds.DaemonSet {
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
func createConfigMap(ctx context.Context, conf *v1.ConfigMap) (*v1.ConfigMap, error) {
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
	_, err2 := api.CoreV1().ConfigMaps(conf.ObjectMeta.Namespace).Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		_, err := api.CoreV1().ConfigMaps(conf.ObjectMeta.Namespace).Create(ctx, conf, listOpts)
		if err != nil {
			logrus.Errorf("Error create configmap: %v", err2)
			return nil, err
		}
	}
	return nil, nil
}
func createServiceAccount(ctx context.Context, conf *v1.ServiceAccount) (*v1.ServiceAccount, error) {
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
	_, err2 := api.CoreV1().ServiceAccounts(conf.ObjectMeta.Namespace).Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		_, err := api.CoreV1().ServiceAccounts(conf.ObjectMeta.Namespace).Create(ctx, conf, listOpts)
		if err != nil {
			logrus.Errorf("Error create serviceAccount: %v", err1)
			return nil, err
		}
	}
	return nil, nil
}
func createClusterRole(ctx context.Context, conf *rbac.ClusterRole) (*rbac.ClusterRole, error) {
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
	_, err2 := api.RbacV1().ClusterRoles().Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		_, err := api.RbacV1().ClusterRoles().Create(ctx, conf, listOpts)
		if err != nil {
			logrus.Errorf("Error create clusterRole: %v", err1)
			return nil, err
		}
	}
	return nil, nil
}
func createClusterRoleBinding(ctx context.Context, conf *rbac.ClusterRoleBinding) (*rbac.ClusterRoleBinding, error) {
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
	_, err2 := api.RbacV1().ClusterRoleBindings().Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		_, err := api.RbacV1().ClusterRoleBindings().Create(ctx, conf, listOpts)
		if err != nil {
			logrus.Errorf("Error create clusterRole: %v", err1)
			return nil, err
		}
	}
	return nil, nil
}
func createDaemonSet(ctx context.Context, conf *ds.DaemonSet) (*ds.DaemonSet, error) {
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
	_, err2 := api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Get(ctx, conf.ObjectMeta.Name, getOpts)
	if err2 != nil {
		_, _ = api.AppsV1().DaemonSets(conf.ObjectMeta.Namespace).Create(ctx, conf, listOpts)
		return nil, err2
	}
	return nil, nil
}
