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

package main

import (
	"log"
	"os"
	"sort"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

type byLastProbeTime []batch.JobCondition

func (b byLastProbeTime) Len() int {
	return len(b)
}

func (b byLastProbeTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byLastProbeTime) Less(i, j int) bool {
	return b[i].LastProbeTime.After(b[j].LastProbeTime.Time)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: check-job-status JOB-NAME")
	}

	jobName := os.Args[1]

	config := &restclient.Config{
		Host: "http://127.0.0.1:8080",
	}

	client, err := client.New(config)
	if err != nil {
		log.Fatal(err)
	}

	job, err := client.Extensions().Jobs(api.NamespaceDefault).Get(jobName)
	if err != nil {
		log.Fatal(err)
	}

	conds := job.Status.Conditions
	sort.Sort(byLastProbeTime(conds))

	if len(conds) == 0 {
		log.Printf("Job Status: undefined")
	} else {
		log.Printf("Job Status: %v:%v\n", conds[0].Type, conds[0].Status)
	}
}
