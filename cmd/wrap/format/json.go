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

func (f Json) ExtractKind(k8sObject string) (string, error) {
	var kind KindExtractor
	err := json.Unmarshal([]byte(k8sObject), &kind)
	return strings.ToLower(kind.Kind), err
}

func (f Json) Wrap(k8sObject, name string) (string, error) {
	base := `{
    "apiVersion": "appcontroller.k8s2/v1alpha1",
    "kind": "Definition",
    "metadata": {
        "name": "` + name + `"
    },` + "\n"

	kind, err := f.ExtractKind(k8sObject)
	if err != nil {
		return "", err
	}
	return base + `    "` + kind + `": ` + strings.TrimLeft(k8sObject, " ") + "}\n", nil
}

func (f Json) IndentLevel() int {
	return 4
}
