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
	"log"
	"os"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/kubernetes/typed/apps/v1alpha1"
	batchv1 "k8s.io/client-go/1.5/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/rest"
)

// Interface is as an interface for k8s clients. It expands native k8s client interface.
type Interface interface {
	ConfigMaps() corev1.ConfigMapInterface
	Secrets() corev1.SecretInterface
	Pods() corev1.PodInterface
	Jobs() batchv1.JobInterface
	Services() corev1.ServiceInterface
	ReplicaSets() v1beta1.ReplicaSetInterface
	PetSets() v1alpha1.PetSetInterface
	DaemonSets() v1beta1.DaemonSetInterface
	Deployments() v1beta1.DeploymentInterface
	PersistentVolumeClaims() corev1.PersistentVolumeClaimInterface

	Dependencies() DependenciesInterface
	ResourceDefinitions() ResourceDefinitionsInterface
}

type client struct {
	Clientset *kubernetes.Clientset
	Deps      DependenciesInterface
	ResDefs   ResourceDefinitionsInterface
	namespace string
}

var _ Interface = &client{}

// Dependencies returns dependency client for ThirdPartyResource created by AppController
func (c client) Dependencies() DependenciesInterface {
	return c.Deps
}

// ResourceDefinitions returns resource definition client for ThirdPartyResource created by AppController
func (c client) ResourceDefinitions() ResourceDefinitionsInterface {
	return c.ResDefs
}

// ConfigMaps returns K8s ConfigMaps client for ac namespace
func (c client) ConfigMaps() corev1.ConfigMapInterface {
	return c.Clientset.ConfigMaps(c.namespace)
}

// Secrets returns K8s Secrets client for ac namespace
func (c client) Secrets() corev1.SecretInterface {
	return c.Clientset.Secrets(c.namespace)
}

// Pods returns K8s Pod client for ac namespace
func (c client) Pods() corev1.PodInterface {
	return c.Clientset.Pods(c.namespace)
}

// Jobs returns K8s Job client for ac namespace
func (c client) Jobs() batchv1.JobInterface {
	return c.Clientset.Batch().Jobs(c.namespace)
}

// Services returns K8s Service client for ac namespace
func (c client) Services() corev1.ServiceInterface {
	return c.Clientset.Services(c.namespace)
}

// ReplicaSets returns K8s ReplicaSet client for ac namespace
func (c client) ReplicaSets() v1beta1.ReplicaSetInterface {
	return c.Clientset.Extensions().ReplicaSets(c.namespace)
}

// PetSets returns K8s PetSet client for ac namespace
func (c client) PetSets() v1alpha1.PetSetInterface {
	return c.Clientset.Apps().PetSets(c.namespace)
}

// DaemonSets return K8s DaemonSet client for ac namespace
func (c client) DaemonSets() v1beta1.DaemonSetInterface {
	return c.Clientset.Extensions().DaemonSets(c.namespace)
}

// Deployments return K8s Deployment client for ac namespace
func (c client) Deployments() v1beta1.DeploymentInterface {
	return c.Clientset.Extensions().Deployments(c.namespace)
}

// PersistentVolumeClaims return K8s PVC client for ac namespace
func (c client) PersistentVolumeClaims() corev1.PersistentVolumeClaimInterface {
	return c.Clientset.PersistentVolumeClaims(c.namespace)
}

func newForConfig(c rest.Config) (Interface, error) {
	deps, err := newDependencies(c)
	if err != nil {
		return nil, err
	}
	resdefs, err := newResourceDefinitions(c)
	if err != nil {
		return nil, err
	}
	cl, err := kubernetes.NewForConfig(&c)
	if err != nil {
		return nil, err
	}

	return &client{
		Clientset: cl,
		Deps:      deps,
		ResDefs:   resdefs,
		namespace: getNamespace(),
	}, nil
}

func thirdPartyResourceRESTClient(c *rest.Config) (*rest.RESTClient, error) {
	c.APIPath = "/apis"
	c.ContentConfig = rest.ContentConfig{
		GroupVersion: &unversioned.GroupVersion{
			Group:   "appcontroller.k8s",
			Version: "v1alpha1",
		},
		NegotiatedSerializer: api.Codecs,
	}
	return rest.RESTClientFor(c)
}

// GetConfig returns restclient.Config for given URL.
// If url is empty, assume in-cluster config. Otherwise, return config for remote cluster.
func GetConfig(url string) (*rest.Config, error) {
	if url == "" {
		log.Println("No Kubernetes cluster URL provided. Assume in-cluster.")
		return rest.InClusterConfig()

	}
	return &rest.Config{Host: url}, nil
}

// New returns client k8s api server under given url
func New(url string) (Interface, error) {
	rc, err := GetConfig(url)
	if err != nil {
		return nil, err
	}
	return newForConfig(*rc)
}

func getNamespace() string {
	ns := os.Getenv("KUBERNETES_AC_POD_NAMESPACE")
	if ns == "" {
		ns = api.NamespaceDefault
	}
	return ns
}
