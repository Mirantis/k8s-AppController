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

package resources

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/copier"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/labels"
)

func init() {
	factories := [...]interfaces.ResourceTemplate{
		configMapTemplateFactory{},
		daemonSetTemplateFactory{},
		deploymentTemplateFactory{},
		flowTemplateFactory{},
		jobTemplateFactory{},
		persistentVolumeClaimTemplateFactory{},
		petSetTemplateFactory{},
		podTemplateFactory{},
		replicaSetTemplateFactory{},
		secretTemplateFactory{},
		serviceTemplateFactory{},
		serviceAccountTemplateFactory{},
		statefulSetTemplateFactory{},
	}
	for _, factory := range factories {
		KindToResourceTemplate[factory.Kind()] = factory
	}
}

// KindToResourceTemplate is a map mapping kind strings to empty structs representing proper resources
// structs implement interfaces.ResourceTemplate
var KindToResourceTemplate = map[string]interfaces.ResourceTemplate{}

// Kinds is slice of keys from KindToResourceTemplate
var Kinds = getKeys(KindToResourceTemplate)

func getKeys(m map[string]interfaces.ResourceTemplate) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func resourceListProgress(resources []interfaces.Resource) (float32, error) {
	var total float32
	var lastErr error
	if len(resources) == 0 {
		return 1, nil
	}
	for _, r := range resources {
		progress, err := r.GetProgress()
		if err == nil {
			total += progress
		} else {
			lastErr = err
		}
	}
	if total > 0 {
		lastErr = nil
	}
	total /= float32(len(resources))
	return total, lastErr
}

func checkExistence(r interfaces.Resource) bool {
	log.Println("Looking for", r.Key())
	_, err := r.GetProgress()
	return err == nil

}

func createExistingResource(r interfaces.Resource) error {
	if !checkExistence(r) {
		log.Printf("Expected resource %s to exist, not found", r.Key())
		return errors.New("resource not found")
	}
	return nil
}

func podsStateFromLabels(apiClient client.Interface, objLabels map[string]string) (float32, error) {
	var labelSelectors []string
	for k, v := range objLabels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
	}
	stringSelector := strings.Join(labelSelectors, ",")
	selector, err := labels.Parse(stringSelector)
	if err != nil {
		return 0, err
	}
	options := v1.ListOptions{LabelSelector: selector.String()}

	pods, err := apiClient.Pods().List(options)
	if err != nil {
		return 0, err
	}
	resources := make([]interfaces.Resource, 0, len(pods.Items))
	for _, pod := range pods.Items {
		p := pod
		resources = append(resources, createNewPod(&p, apiClient.Pods()))
	}

	return resourceListProgress(resources)
}

// Substitutes flow arguments into resource structure. Returns modified copy of the resource
func parametrizeResource(resource interface{}, context interfaces.GraphContext, replaceIn []string) interface{} {
	return copier.CopyWithReplacements(resource, func(p string) string {
		value := context.GetArg(p)
		if value == "" {
			return "$" + p
		}
		return value
	}, append(replaceIn, "ObjectMeta")...)
}
