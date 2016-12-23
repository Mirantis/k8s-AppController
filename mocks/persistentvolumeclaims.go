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
	"strings"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

// MakePersistentVolumeClaim creates a persistentVolumeClaim based on its name
func MakePersistentVolumeClaim(name string) *v1.PersistentVolumeClaim {
	phase := strings.Split(name, "-")[0]
	pvc := &v1.PersistentVolumeClaim{}
	pvc.Name = name
	pvc.Namespace = "testing"
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
