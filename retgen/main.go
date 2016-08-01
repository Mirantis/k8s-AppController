package main

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"
)

var scheme *runtime.Scheme = runtime.NewScheme()
var codec runtime.ParameterCodec

func GetAppControllerClient() (*AppControllerClient, error) {
	version := unversioned.GroupVersion{
		Version: "v1alpha1",
	}

	scheme.AddUnversionedTypes(version, &DependencyList{})
	scheme.AddUnversionedTypes(version, &Dependency{})
	codec = runtime.NewParameterCodec(scheme)

	config := &restclient.Config{
		Host:    "http://localhost:8800",
		APIPath: "/apis/appcontroller.k8s",
		ContentConfig: restclient.ContentConfig{
			GroupVersion:         &version,
			NegotiatedSerializer: api.Codecs,
		},
	}
	client, err := New(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func main() {

}
