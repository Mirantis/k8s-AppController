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
	"strconv"
	"strings"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/labels"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
)

// Base is a base struct that contains data common for all resources
type Base struct {
	meta map[string]interface{}
}

// Meta returns metadata parameter with given name, or empty string,
// if no metadata were provided or such parameter does not exist.
func (b Base) Meta(paramName string) interface{} {
	if b.meta == nil {
		return nil
	}

	val, ok := b.meta[paramName]

	if !ok {
		return nil
	}

	return val
}

// KindToResourceTemplate is a map mapping kind strings to empty structs representing proper resources
// structs implement interfaces.ResourceTemplate
var KindToResourceTemplate = map[string]interfaces.ResourceTemplate{
	"daemonset":             DaemonSet{},
	"job":                   Job{},
	"statefulset":           StatefulSet{},
	"petset":                PetSet{},
	"pod":                   Pod{},
	"replicaset":            ReplicaSet{},
	"service":               Service{},
	"configmap":             ConfigMap{},
	"secret":                Secret{},
	"deployment":            Deployment{},
	"persistentvolumeclaim": PersistentVolumeClaim{},
	"serviceaccount":        ServiceAccount{},
}

// Kinds is slice of keys from KindToResourceTemplate
var Kinds = getKeys(KindToResourceTemplate)

func getKeys(m map[string]interfaces.ResourceTemplate) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func resourceListReady(resources []interfaces.BaseResource) (string, error) {
	for _, r := range resources {
		log.Printf("Checking status for resource %s", r.Key())
		status, err := r.Status(nil)
		if err != nil {
			return "error", err
		}
		if status != "ready" {
			return "not ready", fmt.Errorf("Resource %s is not ready", r.Key())
		}
	}
	return "ready", nil
}

func getPercentage(factorName string, meta map[string]string) (int32, error) {
	var factor string
	var ok bool
	if meta == nil {
		factor = "100"
	} else if factor, ok = meta[factorName]; !ok {
		factor = "100"
	}

	f, err := strconv.ParseInt(factor, 10, 32)
	if (f < 0 || f > 100) && err == nil {
		err = fmt.Errorf("%s factor not between 0 and 100", factorName)
	}
	return int32(f), err
}

func checkExistence(r interfaces.BaseResource) error {
	log.Println("Looking for ", r.Key())
	status, err := r.Status(nil)

	if err == nil {
		log.Printf("Found %s, status: %s ", r.Key(), status)
		return nil
	}

	return err
}

func createExistingResource(r interfaces.BaseResource) error {
	if err := checkExistence(r); err != nil {
		log.Printf("Expected resource %s to exist, not found", r.Key())
		return errors.New("Resource not found")
	}
	return nil
}

func podsStateFromLabels(apiClient client.Interface, objLabels map[string]string) (string, error) {
	var labelSelectors []string
	for k, v := range objLabels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
	}
	stringSelector := strings.Join(labelSelectors, ",")
	selector, err := labels.Parse(stringSelector)
	if err != nil {
		return "error", err
	}
	options := v1.ListOptions{LabelSelector: selector.String()}

	pods, err := apiClient.Pods().List(options)
	if err != nil {
		return "error", err
	}
	resources := make([]interfaces.BaseResource, 0, len(pods.Items))
	for _, pod := range pods.Items {
		p := pod
		resources = append(resources, NewPod(&p, apiClient.Pods(), nil))
	}

	status, err := resourceListReady(resources)
	if status != "ready" || err != nil {
		return status, err
	}

	return "ready", nil
}

// GetIntMeta returns metadata value for parameter 'paramName', or 'defaultValue'
// if parameter is not set or is not an integer value
func GetIntMeta(r interfaces.BaseResource, paramName string, defaultValue int) int {
	value := r.Meta(paramName)
	if value == nil {
		return defaultValue
	}

	intVal, ok := value.(float64)
	if !ok {
		log.Printf("Metadata parameter '%s' for resource '%s' is set to '%v' but it does not seem to be a number, using default value %d", paramName, r.Key(), value, defaultValue)
		return defaultValue
	}

	return int(intVal)
}
