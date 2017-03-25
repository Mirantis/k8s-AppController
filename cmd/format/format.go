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

import "strings"

// Format is an interface for data formats for wrapper
type Format interface {
	ExtractData(k8sObject string) (DataExtractor, error)
	Wrap(k8sObject string) (string, error)
	IndentLevel() int
}

// DataExtractor is a type for extracting data relevant for wrap tool from serialized k8s objects
type DataExtractor struct {
	Kind     string "kind"
	Metadata struct {
		Name string "name"
		GenerateName string "generateName"
	} "metadata"
}

func normalizeName(kind, name string) string {
	if name == "" {
		return ""
	}
	for k, v := range map[string]string{"$": "", "_": "-"} {
		name = strings.Replace(name, k, v, -1)
	}
	return kind + "-" + strings.ToLower(name)
}
