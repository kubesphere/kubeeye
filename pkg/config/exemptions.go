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

import "strings"

func (conf Configuration) IsActionable(ruleID, controllerName string) bool {
	if severity, ok := conf.Checks[ruleID]; !ok || !severity.IsActionable() {
		return false
	}
	if conf.DisallowExemptions {
		return true
	}

	for _, example := range conf.Exemptions {
		for _, rule := range example.Rules {
			if rule != ruleID {
				continue
			}
			for _, controller := range example.ControllerNames {
				if strings.HasPrefix(controllerName, controller) {
					return false
				}
			}
		}
		if len(example.Rules) == 0 {
			for _, controller := range example.ControllerNames {
				if strings.HasPrefix(controllerName, controller) {
					return false
				}
			}
		}
	}
	return true
}
