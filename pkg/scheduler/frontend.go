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
	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
	"k8s.io/client-go/pkg/labels"
)

type Scheduler struct {
	client      client.Interface
	selector    labels.Selector
	concurrency int
}

func New(client client.Interface, selector labels.Selector, concurrency int) interfaces.Scheduler {
	return &Scheduler{
		client:      client,
		selector:    selector,
		concurrency: concurrency,
	}
}
