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

package mocks

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type persistentVolumeClaimClient struct {
}

// MakePersistentVolumeClaim creates a persistentVolumeClaim based on its name
func MakePersistentVolumeClaim(name string) *api.PersistentVolumeClaim {
	phase := strings.Split(name, "-")[0]
	pvc := &api.PersistentVolumeClaim{}
	pvc.Name = name
	switch phase {
	case string(api.ClaimPending):
		pvc.Status.Phase = api.ClaimPending
	case string(api.ClaimBound):
		pvc.Status.Phase = api.ClaimBound
	case string(api.ClaimLost):
		pvc.Status.Phase = api.ClaimLost
	default:
		pvc.Status.Phase = api.ClaimBound
	}

	return pvc
}

func (s *persistentVolumeClaimClient) List(opts api.ListOptions) (*api.PersistentVolumeClaimList, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Get(name string) (*api.PersistentVolumeClaim, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock persistentVolumeClaim %s returned error", name)
	}

	return MakePersistentVolumeClaim(name), nil
}

func (s *persistentVolumeClaimClient) Create(srv *api.PersistentVolumeClaim) (*api.PersistentVolumeClaim, error) {
	return MakePersistentVolumeClaim(srv.Name), nil
}

func (s *persistentVolumeClaimClient) Update(srv *api.PersistentVolumeClaim) (*api.PersistentVolumeClaim, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) UpdateStatus(srv *api.PersistentVolumeClaim) (*api.PersistentVolumeClaim, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Delete(name string) error {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) restclient.ResponseWrapper {
	panic("not implemented")
}

func NewPersistentVolumeClaimClient() unversioned.PersistentVolumeClaimInterface {
	return &persistentVolumeClaimClient{}
}
