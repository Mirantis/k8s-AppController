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
	"fmt"
	"log"
	"strconv"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

type ReplicaSet struct {
	ReplicaSet *extensions.ReplicaSet
	Client     unversioned.ReplicaSetInterface
	Meta       map[string]string
}

func getSuccessFactor(meta map[string]string) (int32, error) {
	var successFactor string
	var ok bool
	if meta == nil {
		successFactor = "100"
	} else if successFactor, ok = meta["success_factor"]; !ok {
		successFactor = "100"
	}

	sf, err := strconv.ParseInt(successFactor, 10, 32)
	// TODO: check 0 < rs <= 100
	return int32(sf), err
}

func replicaSetStatus(r unversioned.ReplicaSetInterface, name string, meta map[string]string) (string, error) {
	rs, err := r.Get(name)
	if err != nil {
		return "error", err
	}

	successFactor, err := getSuccessFactor(meta)
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

func (r ReplicaSet) UpdateMeta(meta map[string]string) error {
	for key, value := range meta {
		oldValue, exists := r.Meta[key]
		if exists && oldValue != value {
			return fmt.Errorf("Resource %v already has meta %v set to %v. Can not set it to %v",
				r, key, oldValue, value)
		}
		r.Meta[key] = value
	}
	return nil
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

func (r ReplicaSet) Status(meta map[string]string) (string, error) {
	return replicaSetStatus(r.Client, r.ReplicaSet.Name, r.Meta)
}

func NewReplicaSet(replicaSet *extensions.ReplicaSet, client unversioned.ReplicaSetInterface) ReplicaSet {
	return ReplicaSet{ReplicaSet: replicaSet, Client: client, Meta: make(map[string]string)}
}

type ExistingReplicaSet struct {
	Name   string
	Client unversioned.ReplicaSetInterface
	Meta   map[string]string
}

func (r ExistingReplicaSet) UpdateMeta(meta map[string]string) error {
	for key, value := range meta {
		oldValue, exists := r.Meta[key]
		if exists && oldValue != value {
			return fmt.Errorf("Resource %v already has meta %v set to %v. Can not set it to %v",
				r, key, oldValue, value)
		}
		r.Meta[key] = value
	}
	return nil
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
	return replicaSetStatus(r.Client, r.Name, r.Meta)
}

func NewExistingReplicaSet(name string, client unversioned.ReplicaSetInterface) ExistingReplicaSet {
	return ExistingReplicaSet{Name: name, Client: client, Meta: make(map[string]string)}
}
