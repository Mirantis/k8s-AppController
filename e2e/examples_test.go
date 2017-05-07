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
	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Examples Suite", func() {
	options := interfaces.DependencyGraphOptions{ReplicaCount: 1}
	framework := ExamplesFramework{testutils.NewAppControllerManager()}

	It("Example 'simple' should finish", func() {
		framework.CreateRunAndVerify("simple", options)
	})

	It("Example 'services' should finish", func() {
		framework.CreateRunAndVerify("services", options)
	})

	It("Example 'extended' should finish", func() {
		testutils.SkipIf14()
		framework.CreateRunAndVerify("extended", options)
	})
})
