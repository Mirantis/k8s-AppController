// Copyright 2017 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mirantis/k8s-AppController/cmd/format"
	"github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
	"github.com/Mirantis/k8s-AppController/pkg/scheduler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
)

type examplesFramework struct {
	*utils.AppControllerManager
}

func (f *examplesFramework) CreateExample(exampleName string) {
	// list all files in directory, for each file run Create
	fqDir := filepath.Join(utils.TestContext.Examples, exampleName)
	By("Creating example from directory " + fqDir)
	files, err := ioutil.ReadDir(fqDir)
	Expect(err).NotTo(HaveOccurred())
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			f.Create(filepath.Join(utils.TestContext.Examples, exampleName, file.Name()))
		}
	}
}

func (f *examplesFramework) Create(fileName string) {
	// Load data from filepath
	// try to serialize it into Unstructured or UnstructuredList
	data, err := ioutil.ReadFile(fileName)
	Expect(err).NotTo(HaveOccurred())

	utils.ForEachYamlDocument(data, func(doc []byte) {
		jsonData, err := utilyaml.ToJSON(doc)
		Expect(err).NotTo(HaveOccurred())
		var ust runtime.Unstructured
		err = runtime.DecodeInto(api.Codecs.UniversalDecoder(), jsonData, &ust)
		if err != nil {
			var ustList runtime.UnstructuredList
			err = runtime.DecodeInto(api.Codecs.UniversalDecoder(), jsonData, &ustList)
			Expect(err).NotTo(HaveOccurred())
			f.handleListCreation(&ustList)
			return
		}
		f.handleItemCreation(&ust)
	})
}

func (f *examplesFramework) handleItemCreation(ust *runtime.Unstructured) {
	// based on kind of item - instantiate correct object and wrap it with resource definition
	// if unstructured is of kind Dependency - just create it as is
	encodedData, err := json.Marshal(ust)
	Expect(err).NotTo(HaveOccurred())
	utils.Logf("Found resource with name %v and kind %v\n", ust.GetName(), ust.GetKind())
	if ust.GetKind() == "Dependency" {
		var dep client.Dependency
		err = json.Unmarshal(encodedData, &dep)
		Expect(err).NotTo(HaveOccurred())
		_, err = f.Client.Dependencies().Create(&dep)
		Expect(err).NotTo(HaveOccurred())
		return
	}
	formatter := format.JSON{}
	wrapped, err := formatter.Wrap(string(encodedData))
	Expect(err).NotTo(HaveOccurred())
	var resDef client.ResourceDefinition
	err = json.Unmarshal([]byte(wrapped), &resDef)
	Expect(err).NotTo(HaveOccurred())
	_, err = f.Client.ResourceDefinitions().Create(&resDef)
	Expect(err).NotTo(HaveOccurred())
}

func (f *examplesFramework) handleListCreation(ustList *runtime.UnstructuredList) {
	for _, ust := range ustList.Items {
		f.handleItemCreation(ust)
	}
}

func (f *examplesFramework) VerifyStatus(options interfaces.DependencyGraphOptions) {
	var depReport report.DeploymentReport
	Eventually(
		func() bool {
			status, r, err := scheduler.GetStatus(f.Client, nil, options)
			if err != nil {
				return false
			}
			depReport = r.(report.DeploymentReport)
			utils.Logf("STATUS: %s\n", status)
			return status == interfaces.Finished || status == interfaces.Empty
		},
		300*time.Second, 5*time.Second).Should(BeTrue(), strings.Join(depReport.AsText(0), "\n"))
}

func (f *examplesFramework) CreateRunAndVerify(exampleName string, options interfaces.DependencyGraphOptions) {
	By("Creating example " + exampleName)
	f.CreateExample(exampleName)
	By("Running appcontroller scheduler")
	f.RunWithOptions(false, options)
	By("Verifying status of deployment for example " + exampleName)
	f.VerifyStatus(options)
}
