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
	"github.com/stretchr/testify/assert"
	"testing"
)

var confInvalid = `test`

var confValidYaml = `
checks:
  cpuLimitsMissing: warning
`

func TestParseError(t *testing.T) {
	_, err := Parse([]byte(confInvalid))
	expectedErr := "Decoding config failed: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Configuration"
	assert.EqualError(t, err, expectedErr)
}

func TestParseYaml(t *testing.T) {
	parsedConf, err := Parse([]byte(confValidYaml))
	assert.NoError(t, err, "Expected no error when parsing YAML config")
	testParsedConfig(t, &parsedConf)
}

func TestConfigFrom(t *testing.T) {
	var parsedConf Configuration
	var err error
	parsedConf, err = ParseFile1("../../examples/tmp/rule.yaml")
	assert.NoError(t, err)
	testParsedFileConfig(t, &parsedConf)
}

func testParsedConfig(t *testing.T, config *Configuration) {
	assert.Equal(t, SeverityWarning, config.Checks["cpuLimitsMissing"])
}
func testParsedFileConfig(t *testing.T, config *Configuration) {
	assert.Equal(t, SeverityWarning, config.Checks["imageFromUnauthorizedRegistry"])
}
