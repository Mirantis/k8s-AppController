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

package client

import (
	"bytes"
	"encoding/json"

	"github.com/Mirantis/k8s-AppController/pkg/client/petsets/apis/apps/v1alpha1"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	appsbeta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

// ResourceDefinition represents template for one of supported k8s resource types
type ResourceDefinition struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard object metadata
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Meta map[string]interface{} `json:"meta,omitempty"`

	//TODO: add other object types
	Pod                   *v1.Pod                   `json:"pod,omitempty"`
	Job                   *batchv1.Job              `json:"job,omitempty"`
	Service               *v1.Service               `json:"service,omitempty"`
	ReplicaSet            *v1beta1.ReplicaSet       `json:"replicaset,omitempty"`
	StatefulSet           *appsbeta1.StatefulSet    `json:"statefulset,omitempty"`
	ServiceAccount        *v1.ServiceAccount        `json:"serviceaccount,omitempty"`
	PetSet                *v1alpha1.PetSet          `json:"petset,omitempty"`
	DaemonSet             *v1beta1.DaemonSet        `json:"daemonset,omitempty"`
	ConfigMap             *v1.ConfigMap             `json:"configmap,omitempty"`
	Secret                *v1.Secret                `json:"secret,omitempty"`
	Deployment            *v1beta1.Deployment       `json:"deployment, omitempty"`
	PersistentVolumeClaim *v1.PersistentVolumeClaim `json:"persistentvolumeclaim, omitempty"`
	Flow                  *Flow                     `json:"flow, omitempty"`
}

// ResourceDefinitionList is a k8s object representing list of resource definitions
type ResourceDefinitionList struct {
	unversioned.TypeMeta `json:",inline"`

	// Standard list metadata.
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ResourceDefinition `json:"items"`
}

// ResourceDefinitionsInterface is an interface to access Definition objects in k8s
type ResourceDefinitionsInterface interface {
	Create(*ResourceDefinition) (*ResourceDefinition, error)
	List(opts api.ListOptions) (*ResourceDefinitionList, error)
	Delete(name string, opts *api.DeleteOptions) error
}

type resourceDefinitions struct {
	rc        *rest.RESTClient
	namespace string
}

// GetObjectKind returns type header of ResourceDefinition object
func (r *ResourceDefinition) GetObjectKind() unversioned.ObjectKind {
	return &r.TypeMeta
}

// GetObjectMeta returns metadata of ResourceDefinition object
func (r *ResourceDefinition) GetObjectMeta() meta.Object {
	return &r.ObjectMeta
}

func newResourceDefinitions(c rest.Config, ns string) (*resourceDefinitions, error) {
	rc, err := thirdPartyResourceRESTClient(&c)
	if err != nil {
		return nil, err
	}

	return &resourceDefinitions{rc, ns}, nil
}

// List returns Definition objects stored in K8s
func (c *resourceDefinitions) List(opts api.ListOptions) (*ResourceDefinitionList, error) {
	resp, err := c.rc.Get().
		Namespace(c.namespace).
		Resource("definitions").
		LabelsSelectorParam(opts.LabelSelector).
		DoRaw()

	if err != nil {
		return nil, err
	}

	result := &ResourceDefinitionList{}
	err = json.NewDecoder(bytes.NewReader(resp)).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create creates new Definition object in K8s
func (c *resourceDefinitions) Create(rd *ResourceDefinition) (result *ResourceDefinition, err error) {
	result = &ResourceDefinition{}
	err = c.rc.Post().
		Resource("definitions").
		Namespace(c.namespace).
		Body(rd).
		Do().
		Into(result)
	return
}

// Delete deletes Definition object from K8s
func (c *resourceDefinitions) Delete(name string, opts *api.DeleteOptions) error {
	return c.rc.Delete().
		Namespace(c.namespace).
		Resource("definitions").
		Name(name).
		Body(opts).
		Do().
		Error()
}
