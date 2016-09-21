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
	"encoding/json"
	"strings"
)

type Json struct {
}

// ExtractData returns data relevant for wrap tool from serialized k8s object
func (f Json) ExtractData(k8sObject string) (DataExtractor, error) {
	var data DataExtractor
	err := json.Unmarshal([]byte(k8sObject), &data)
	data.Kind = strings.ToLower(data.Kind)
	return data, err
}

// Wrap wraps k8sObject into Definition ThirdPArtyResource
func (f Json) Wrap(k8sObject string) (string, error) {
	data, err := f.ExtractData(k8sObject)

	base := `{
    "apiVersion": "appcontroller.k8s2/v1alpha1",
    "kind": "Definition",
    "metadata": {
        "name": "` + data.Kind + "-" + data.Metadata.Name + `"
    },` + "\n"

	if err != nil {
		return "", err
	}
	return base + `    "` + data.Kind + `": ` + strings.TrimLeft(k8sObject, " ") + "}\n", nil
}

func (f Json) IndentLevel() int {
	return 4
}
