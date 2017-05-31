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
	appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	extbeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/labels"

	"github.com/Mirantis/k8s-AppController/pkg/client"
	appsalpha1 "github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"
	"github.com/Mirantis/k8s-AppController/pkg/copier"
	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
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

// Base is a base struct that contains data common for all resources
type Base struct {
	Definition client.ResourceDefinition
	meta       map[string]interface{}
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

// equalsToDefinition returns false since there is no definition to compare Base resource against.
func (b Base) equalsToDefinition(_ interface{}) bool {
	return false
}

// UpdateFromDefinition is an empty implementation
func (b Base) UpdateFromDefinition() error {
	return nil
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

func resourceListStatus(resources []interfaces.BaseResource) (interfaces.ResourceStatus, error) {
	for _, r := range resources {
		status, err := r.Status(nil)
		if err != nil {
			return interfaces.ResourceError, err
		}
		if status != interfaces.ResourceReady {
			return interfaces.ResourceNotReady, fmt.Errorf("resource %s is not ready", r.Key())
		}
	}
	return interfaces.ResourceReady, nil
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
	log.Println("Looking for", r.Key())
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
		return errors.New("resource not found")
	}
	return nil
}

func podsStateFromLabels(apiClient client.Interface, objLabels map[string]string) (interfaces.ResourceStatus, error) {
	var labelSelectors []string
	for k, v := range objLabels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
	}
	stringSelector := strings.Join(labelSelectors, ",")
	selector, err := labels.Parse(stringSelector)
	if err != nil {
		return interfaces.ResourceError, err
	}
	options := v1.ListOptions{LabelSelector: selector.String()}

	pods, err := apiClient.Pods().List(options)
	if err != nil {
		return interfaces.ResourceError, err
	}
	resources := make([]interfaces.BaseResource, 0, len(pods.Items))
	for _, pod := range pods.Items {
		p := pod
		//resources = append(resources, createNewPod(&p, apiClient.Pods(), nil))
		def := MakeDefinition(&p)
		resources = append(resources, createNewPod(def, apiClient.Pods()))
	}

	status, err := resourceListStatus(resources)
	if status != interfaces.ResourceReady || err != nil {
		return status, err
	}

	return interfaces.ResourceReady, nil
}

// GetIntMeta returns metadata value for parameter 'paramName', or 'defaultValue'
// if parameter is not set or is not an integer value
func GetIntMeta(r interfaces.BaseResource, paramName string, defaultValue int) int {
	value := r.Meta(paramName)
	if value == nil {
		return defaultValue
	}

	intVal, ok := value.(int)
	if ok {
		return intVal
	}

	floatVal, ok := value.(float64)

	if !ok {
		log.Printf("Metadata parameter '%s' for resource '%s' is set to '%v' but it does not seem to be a number, using default value %d", paramName, r.Key(), value, defaultValue)
		return defaultValue
	}

	return int(floatVal)
}

// GetStringMeta returns metadata value for parameter 'paramName', or 'defaultValue'
// if parameter is not set or is not a string value
func GetStringMeta(r interfaces.BaseResource, paramName string, defaultValue string) string {
	value := r.Meta(paramName)
	if value == nil {
		return defaultValue
	}

	strVal, ok := value.(string)
	if !ok {
		log.Printf("Metadata parameter '%s' for resource '%s' is set to '%v' but it does not seem to be a string, using default value %s", paramName, r.Key(), value, defaultValue)
		return defaultValue
	}

	return string(strVal)
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

func MakeDefinition(object interface{}) client.ResourceDefinition {
	result := client.ResourceDefinition{}

	switch object.(type) {
	case *v1.ConfigMap:
		result.ConfigMap = object.(*v1.ConfigMap)
	case *extbeta1.DaemonSet:
		result.DaemonSet = object.(*extbeta1.DaemonSet)
	case *extbeta1.Deployment:
		result.Deployment = object.(*extbeta1.Deployment)
	case *batchv1.Job:
		result.Job = object.(*batchv1.Job)
	case *v1.PersistentVolumeClaim:
		result.PersistentVolumeClaim = object.(*v1.PersistentVolumeClaim)
	case *appsalpha1.PetSet:
		result.PetSet = object.(*appsalpha1.PetSet)
	case *v1.Pod:
		result.Pod = object.(*v1.Pod)
	case *extbeta1.ReplicaSet:
		result.ReplicaSet = object.(*extbeta1.ReplicaSet)
	case *v1.Secret:
		result.Secret = object.(*v1.Secret)
	case *v1.Service:
		result.Service = object.(*v1.Service)
	case *v1.ServiceAccount:
		result.ServiceAccount = object.(*v1.ServiceAccount)
	case *appsbeta1.StatefulSet:
		result.StatefulSet = object.(*appsbeta1.StatefulSet)
	}
	return result
}
