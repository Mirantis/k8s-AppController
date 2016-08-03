package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
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
	fmt.Print(getWrappedYaml(definition, os.Args[1]))
}
