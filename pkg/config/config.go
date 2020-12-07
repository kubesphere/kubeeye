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

package config

import (
	"bytes"
	"errors"
	"fmt"
	packr "github.com/gobuffalo/packr/v2"
	"io"
	"k8s.io/apimachinery/pkg/util/yaml"
	"path"
)

type Configuration struct {
	Checks             map[string]Severity    `json:"checks"`
	CustomChecks       map[string]SchemaCheck `json:"customChecks"`
	Exemptions         []Exemption            `json:"exemptions"`
	DisallowExemptions bool                   `json:"disallowExemptions"`
}

type Exemption struct {
	Rules           []string `json:"rules"`
	ControllerNames []string `json:"controllerNames"`
}

var configBox = (*packr.Box)(nil)
var configBox1 = (*packr.Box)(nil)

func getConfigBox() *packr.Box {
	if configBox == (*packr.Box)(nil) {
		configBox = packr.New("Config", "../../examples")

	}
	return configBox
}

func getConfigBox1(fp string) *packr.Box {
	var dir, _ = path.Split(fp)
	if configBox1 == (*packr.Box)(nil) {
		configBox1 = packr.New("CustomConfig", fmt.Sprintf("%s", dir))
	}
	return configBox1
}

func ParseFile() (Configuration, error) {
	var rawBytes []byte
	var err error

	rawBytes, err = getConfigBox().Find("config.yaml")
	if err != nil {
		return Configuration{}, err
	}
	return Parse(rawBytes)
}
func ParseFile1(fp string) (Configuration, error) {
	var rawBytes []byte
	var err error

	rawBytes, err = getConfigBox1(fp).Find("config.yaml")
	if err != nil {
		return Configuration{}, err
	}
	return Parse(rawBytes)
}
func Parse(rawBytes []byte) (Configuration, error) {
	reader := bytes.NewReader(rawBytes)
	conf := Configuration{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&conf); err != nil {
			if err == io.EOF {
				break
			}
			return conf, fmt.Errorf("Decoding config failed: %v", err)
		}
	}
	for key, check := range conf.CustomChecks {
		err := check.Initialize(key)
		if err != nil {
			return conf, err
		}
		conf.CustomChecks[key] = check
		if _, ok := conf.Checks[key]; !ok {
			return conf, fmt.Errorf("no severity specified for custom kubeye %s. Please add the following to your configuration:\n\nchecks:\n  %s: warning # or danger/ignore\n\nto enable your kubeye", key, key)
		}
	}
	return conf, conf.Validate()
}
func (c Configuration) Validate() error {
	if len(c.Checks) == 0 {
		return errors.New("No checks were enabled")
	}
	return nil
}
