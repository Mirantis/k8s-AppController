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

	"k8s.io/client-go/kubernetes"
	appsbeta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/rest"
)

// Interface is as an interface for k8s clients. It expands native k8s client interface.
type Interface interface {
	ConfigMaps() corev1.ConfigMapInterface
	Secrets() corev1.SecretInterface
	Pods() corev1.PodInterface
	Jobs() batchv1.JobInterface
	Services() corev1.ServiceInterface
	ReplicaSets() v1beta1.ReplicaSetInterface
	StatefulSets() appsbeta1.StatefulSetInterface
	DaemonSets() v1beta1.DaemonSetInterface
	Deployments() v1beta1.DeploymentInterface
	PersistentVolumeClaims() corev1.PersistentVolumeClaimInterface

	Dependencies() DependenciesInterface
	ResourceDefinitions() ResourceDefinitionsInterface
}

type Client struct {
	Clientset kubernetes.Interface
	Deps      DependenciesInterface
	ResDefs   ResourceDefinitionsInterface
	Namespace string
}

var _ Interface = &Client{}

// Dependencies returns dependency client for ThirdPartyResource created by AppController
func (c Client) Dependencies() DependenciesInterface {
	return c.Deps
}

// ResourceDefinitions returns resource definition client for ThirdPartyResource created by AppController
func (c Client) ResourceDefinitions() ResourceDefinitionsInterface {
	return c.ResDefs
}

// ConfigMaps returns K8s ConfigMaps client for ac namespace
func (c Client) ConfigMaps() corev1.ConfigMapInterface {
	return c.Clientset.Core().ConfigMaps(c.Namespace)
}

// Secrets returns K8s Secrets client for ac namespace
func (c Client) Secrets() corev1.SecretInterface {
	return c.Clientset.Core().Secrets(c.Namespace)
}

// Pods returns K8s Pod client for ac namespace
func (c Client) Pods() corev1.PodInterface {
	return c.Clientset.Core().Pods(c.Namespace)
}

// Jobs returns K8s Job client for ac namespace
func (c Client) Jobs() batchv1.JobInterface {
	return c.Clientset.Batch().Jobs(c.Namespace)
}

// Services returns K8s Service client for ac namespace
func (c Client) Services() corev1.ServiceInterface {
	return c.Clientset.Core().Services(c.Namespace)
}

// ReplicaSets returns K8s ReplicaSet client for ac namespace
func (c Client) ReplicaSets() v1beta1.ReplicaSetInterface {
	return c.Clientset.Extensions().ReplicaSets(c.Namespace)
}

// StatefulSets returns K8s StatefulSet client for ac namespace
func (c Client) StatefulSets() appsbeta1.StatefulSetInterface {
	return c.Clientset.Apps().StatefulSets(c.Namespace)
}

// DaemonSets return K8s DaemonSet client for ac namespace
func (c Client) DaemonSets() v1beta1.DaemonSetInterface {
	return c.Clientset.Extensions().DaemonSets(c.Namespace)
}

// Deployments return K8s Deployment client for ac namespace
func (c Client) Deployments() v1beta1.DeploymentInterface {
	return c.Clientset.Extensions().Deployments(c.Namespace)
}

// PersistentVolumeClaims return K8s PVC client for ac namespace
func (c Client) PersistentVolumeClaims() corev1.PersistentVolumeClaimInterface {
	return c.Clientset.Core().PersistentVolumeClaims(c.Namespace)
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

	return &Client{
		Clientset: cl,
		Deps:      deps,
		ResDefs:   resdefs,
		Namespace: getNamespace(),
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
