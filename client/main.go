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
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
)

func GetAppControllerClient(url string) (*AppControllerClient, error) {
	version := unversioned.GroupVersion{
		Version: "v1alpha1",
	}

	config := &restclient.Config{
		Host:    url,
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
