// Copyright 2016 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package format

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// Yaml implements format.Format interface
type Yaml struct {
}

// ExtractData returns data relevant for wrap tool from serialized k8s object
func (f Yaml) ExtractData(k8sObject string) (DataExtractor, error) {
	var data DataExtractor
	err := yaml.Unmarshal([]byte(k8sObject), &data)
	data.Kind = strings.ToLower(data.Kind)
	return data, err
}

// Wrap wraps k8sObject into Definition ThirdPArtyResource
func (f Yaml) Wrap(k8sObject string) (string, error) {
	objects := strings.Split(k8sObject, fmt.Sprintf("%s---", strings.Repeat(" ", f.IndentLevel())))

	result := make([]string, 0, len(objects))
	for _, o := range objects {
		data, err := f.ExtractData(o)
		if err != nil {
			return "", err
		}
		base := `apiVersion: appcontroller.k8s/v1alpha1
kind: Definition
metadata:
  name: ` + normalizeName(data.Kind+"-"+data.Metadata.Name) + "\n"
		result = append(result, base+data.Kind+":\n"+strings.Trim(o, "\n"))
	}

	return strings.Join(result, "\n---\n"), nil
}

// IndentLevel returns indent level for Yaml format
func (f Yaml) IndentLevel() int {
	return 2
}
