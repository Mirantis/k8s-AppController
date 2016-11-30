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
	"log"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
)

type ReplicaSet struct {
	ReplicaSet *extensions.ReplicaSet
	Client     unversioned.ReplicaSetInterface
}

func replicaSetStatus(r unversioned.ReplicaSetInterface, name string, meta map[string]string) (string, error) {
	rs, err := r.Get(name)
	if err != nil {
		return "error", err
	}

	successFactor, err := getFactor("success_factor", meta)
	if err != nil {
		return "error", err
	}

	if rs.Status.Replicas*100 < rs.Spec.Replicas*successFactor {
		return "not ready", nil
	}

	return "ready", nil
}

func replicaSetKey(name string) string {
	return "replicaset/" + name
}

func (r ReplicaSet) Key() string {
	return replicaSetKey(r.ReplicaSet.Name)
}

func (r ReplicaSet) Create() error {
	log.Println("Looking for replica set", r.ReplicaSet.Name)
	status, err := r.Status(nil)

	if err == nil {
		log.Printf("Found replica set %s, status: %s ", r.ReplicaSet.Name, status)
		log.Println("Skipping creation of replica set", r.ReplicaSet.Name)
		return nil
	}

	log.Println("Creating replica set", r.ReplicaSet.Name)
	r.ReplicaSet, err = r.Client.Create(r.ReplicaSet)
	return err
}

// Delete deletes ReplicaSet from the cluster
func (r ReplicaSet) Delete() error {
	return r.Client.Delete(r.ReplicaSet.Name, nil)
}

func (r ReplicaSet) Status(meta map[string]string) (string, error) {
	return replicaSetStatus(r.Client, r.ReplicaSet.Name, meta)
}

// NameMatches gets resource definition and a name and checks if
// the ReplicaSet part of resource definition has matching name.
func (r ReplicaSet) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.ReplicaSet != nil && def.ReplicaSet.Name == name
}

// New returns new ReplicaSet based on resource definition
func (r ReplicaSet) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewReplicaSet(def.ReplicaSet, c.ReplicaSets())
}

// NewExisting returns new ExistingReplicaSet based on resource definition
func (r ReplicaSet) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingReplicaSet(name, c.ReplicaSets())
}

func NewReplicaSet(replicaSet *extensions.ReplicaSet, client unversioned.ReplicaSetInterface) ReplicaSet {
	return ReplicaSet{ReplicaSet: replicaSet, Client: client}
}

type ExistingReplicaSet struct {
	Name   string
	Client unversioned.ReplicaSetInterface
	ReplicaSet
}

func (r ExistingReplicaSet) Key() string {
	return replicaSetKey(r.Name)
}

func (r ExistingReplicaSet) Create() error {
	log.Println("Looking for replica set", r.Name)
	status, err := r.Status(nil)

	if err == nil {
		log.Printf("Found replica set %s, status: %s ", r.Name, status)
		log.Println("Skipping creation of replica set", r.Name)
		return nil
	}

	log.Fatalf("Replica set %s not found", r.Name)
	return errors.New("Replica set not found")
}

func (r ExistingReplicaSet) Status(meta map[string]string) (string, error) {
	return replicaSetStatus(r.Client, r.Name, meta)
}

func NewExistingReplicaSet(name string, client unversioned.ReplicaSetInterface) ExistingReplicaSet {
	return ExistingReplicaSet{Name: name, Client: client}
}
