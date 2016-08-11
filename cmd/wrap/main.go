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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func getInput(stream *os.File) string {
	result := ""
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		//add two spaces of identation
		result += "  " + scanner.Text() + "\n"
	}
	return result
}

type KindExtractor struct {
	Kind string "kind"
}

func getKind(k8sObject string) (string, error) {
	var kind KindExtractor
	err := yaml.Unmarshal([]byte(k8sObject), &kind)
	return strings.ToLower(kind.Kind), err
}

func getWrappedYaml(k8sObject, name string) (string, error) {
	base := `apiVersion: appcontroller.k8s2/v1alpha1
kind: Definition
metadata:
  name: ` + name + "\n"

	kind, err := getKind(k8sObject)
	if err != nil {
		return "", err
	}
	return base + kind + ":\n" + k8sObject, nil
}

func main() {
	definition := getInput(os.Stdin)
	out, err := getWrappedYaml(definition, os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
