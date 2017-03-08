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

package resources

import (
	"errors"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"github.com/Mirantis/k8s-AppController/pkg/report"
)

type Flow struct {
	Base
	flow      *client.Flow
	scheduler interfaces.Scheduler
	status    interfaces.ResourceStatus
}

type flowTemplateFactory struct{}

func (flowTemplateFactory) ShortName(definition client.ResourceDefinition) string {
	if definition.Flow == nil {
		return ""
	}
	return definition.Flow.Name
}

func (flowTemplateFactory) Kind() string {
	return "flow"
}

func (flowTemplateFactory) New(def client.ResourceDefinition, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return report.SimpleReporter{
		BaseResource: &Flow{
			Base:      Base{def.Meta},
			flow:      def.Flow,
			scheduler: gc.Scheduler(),
			status:    interfaces.ResourceNotReady,
		}}
}

func (flowTemplateFactory) NewExisting(name string, c client.Interface, gc interfaces.GraphContext) interfaces.Resource {
	return nil
}

func (f Flow) Key() string {
	return "flow/" + f.flow.Name
}

func (f *Flow) Create() error {
	graph, err := f.scheduler.BuildDependencyGraph(f.flow.Name, map[string]string{})
	if err != nil {
		return err
	}
	go func() {
		stopChan := make(chan struct{})
		graph.Deploy(stopChan)
		f.status = interfaces.ResourceReady
	}()
	return nil
}

func (f Flow) Delete() error {
	return errors.New("Not supported yet")
}

func (f Flow) Status(meta map[string]string) (interfaces.ResourceStatus, error) {
	return f.status, nil
}
