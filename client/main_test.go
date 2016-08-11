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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
)

var (
	// testMux is the HTTP request multiplexer used with the test server.
	testMux *http.ServeMux

	// testClient is the JIRA client being tested.
	testClient *AppControllerClient

	// testServer is a test HTTP server used to provide mock API responses.
	testServer *httptest.Server
)

func setup() {
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	version := unversioned.GroupVersion{
		Version: "v1alpha1",
	}

	config := &restclient.Config{
		Host:    testServer.URL,
		APIPath: "/apis/appcontroller.k8s",
		ContentConfig: restclient.ContentConfig{
			GroupVersion:         &version,
			NegotiatedSerializer: api.Codecs,
		},
	}

	testClient, _ = New(config)
}

func teardown() {
	testServer.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testRequestURL(t *testing.T, r *http.Request, want string) {
	if got := r.URL.String(); !strings.HasPrefix(got, want) {
		t.Errorf("Request URL: %v, want %v", got, want)
	}
}

func TestGetDependencies(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/apis/appcontroller.k8s1/v1alpha1/namespaces/default/dependencies"

	raw, err := ioutil.ReadFile("./mocks/dependencies.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(raw))
	})

	deps, err := testClient.Dependencies().List(api.ListOptions{})
	if err != nil {
		t.Error(err.Error())
	}
	if len(deps.Items) != 2 {
		t.Errorf("Wrong dependecy item count. Expected 2, got %d ", len(deps.Items))
	}
}

func TestGetDependency(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/apis/appcontroller.k8s1/v1alpha1/namespaces/default/dependencies/dep-name"

	raw, err := ioutil.ReadFile("./mocks/dependency.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(raw))
	})

	dep, err := testClient.Dependencies().Get("dep-name")
	if err != nil {
		t.Error(err.Error())
	}
	if dep.Child != "pod/test-pod-2" {
		t.Errorf("Wrong dependecy child. Expected `pod/test-pod-2`, got %s ", len(dep.Child))
	}
	if dep.Parent != "pod/test-pod" {
		t.Errorf("Wrong dependecy child. Expected `pod/test-pod`, got %s ", len(dep.Child))
	}
}

func TestGetResourceDefinitions(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/apis/appcontroller.k8s2/v1alpha1/namespaces/default/definitions"

	raw, err := ioutil.ReadFile("./mocks/definitions.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(raw))
	})

	defs, err := testClient.ResourceDefinitions().List(api.ListOptions{})
	if err != nil {
		t.Error(err.Error())
	}
	if len(defs.Items) != 3 {
		t.Errorf("Wrong definitions item count. Expected 3, got %d ", len(defs.Items))
	}
}

func TestGetPodResourceDefinition(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/apis/appcontroller.k8s2/v1alpha1/namespaces/default/definitions/pod-definition-1"

	raw, err := ioutil.ReadFile("./mocks/pod-definition.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(raw))
	})

	def, err := testClient.ResourceDefinitions().Get("pod-definition-1")
	if err != nil {
		t.Error(err.Error())
	}
	if def.Job != nil {
		t.Error("Job should be null")
	}
	if def.Pod == nil {
		t.Error("Pod shouldn't be null")
	}
}

func TestGetJobResourceDefinition(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/apis/appcontroller.k8s2/v1alpha1/namespaces/default/definitions/job-definition-1"

	raw, err := ioutil.ReadFile("./mocks/job-definition.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(raw))
	})

	def, err := testClient.ResourceDefinitions().Get("job-definition-1")
	if err != nil {
		t.Error(err.Error())
	}
	if def.Pod != nil {
		t.Error("Pod should be nil")
	}
	if def.Job == nil {
		t.Error("Job shouldn't be nil")
	}
}

func TestGetURLWithoutOptions(t *testing.T) {
	baseUrl, _ := url.Parse("http://trololo.com:5233")

	noOptionsURL := getUrlWithOptions(baseUrl, api.ListOptions{})

	if noOptionsURL.String() != baseUrl.String() {
		t.Errorf("%s expected, got %s", baseUrl, noOptionsURL)
	}
}

func TestGetURLWithOptions(t *testing.T) {
	baseUrl, _ := url.Parse("http://trololo.com:5233")
	options := api.ListOptions{
		LabelSelector: AppControllerLabelSelector{Key: "Key", Value: "Value"},
	}

	optionsURL := getUrlWithOptions(baseUrl, options)

	expected := fmt.Sprintf("%s?labelSelector=Key%%3DValue", baseUrl.String())

	if optionsURL.String() != expected {
		t.Errorf("%s expected, got %s", expected, optionsURL)
	}
}
