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

	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
)

type persistentVolumeClaimClient struct {
}

// MakePersistentVolumeClaim creates a persistentVolumeClaim based on its name
func MakePersistentVolumeClaim(name string) *v1.PersistentVolumeClaim {
	phase := strings.Split(name, "-")[0]
	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = name
	switch phase {
	case string(api.ClaimPending):
		pvc.Status.Phase = v1.ClaimPending
	case string(api.ClaimBound):
		pvc.Status.Phase = v1.ClaimBound
	case string(api.ClaimLost):
		pvc.Status.Phase = v1.ClaimLost
	default:
		pvc.Status.Phase = v1.ClaimBound
	}

	return pvc
}

func (s *persistentVolumeClaimClient) List(opts api.ListOptions) (*v1.PersistentVolumeClaimList, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Get(name string) (*v1.PersistentVolumeClaim, error) {
	status := strings.Split(name, "-")[0]
	if status == "error" {
		return nil, fmt.Errorf("mock persistentVolumeClaim %s returned error", name)
	}

	return MakePersistentVolumeClaim(name), nil
}

func (s *persistentVolumeClaimClient) Create(srv *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return MakePersistentVolumeClaim(srv.Name), nil
}

func (s *persistentVolumeClaimClient) Update(srv *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) UpdateStatus(srv *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Delete(name string, opts *api.DeleteOptions) error {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Watch(opts api.ListOptions) (watch.Interface, error) {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) ProxyGet(scheme string, name string, port string, path string, params map[string]string) rest.ResponseWrapper {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	panic("not implemented")
}

func (s *persistentVolumeClaimClient) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.PersistentVolumeClaim, err error) {
	panic("not implemented")
}

func NewPersistentVolumeClaimClient() corev1.PersistentVolumeClaimInterface {
	return &persistentVolumeClaimClient{}
}
