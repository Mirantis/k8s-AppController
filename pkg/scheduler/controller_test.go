// Copyright 2017 Mirantis
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

package scheduler

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	fcache "k8s.io/client-go/tools/cache/testing"

	"github.com/Mirantis/k8s-AppController/pkg/mocks"
)

func TestDeploymentWorkflow(t *testing.T) {
	c := mocks.NewClient()
	source := fcache.NewFakeControllerSource()

	waitCh := make(chan struct{})
	var stopCh <-chan struct{}
	f := func(_ DependencyGraph, _ int, stopChan <-chan struct{}) {
		t.Log("Closing wait chanell")
		waitCh <- struct{}{}
		stopCh = stopChan
	}

	controller := newDeploymentController(c, source, f)
	controllerCh := make(chan struct{})
	defer close(controllerCh)
	go controller.Run(controllerCh)

	cfg := &v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{Name: controlName},
		Data: map[string]string{
			concurrencyKey: "0",
			selector:       "",
		},
	}
	source.Add(cfg)
	<-waitCh
	source.Delete(cfg)
	<-stopCh
}
