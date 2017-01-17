// Copyright 2016 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"flag"
	"fmt"
	"io"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Mirantis/k8s-AppController/pkg/client"
)

var CLUSTER_URL string

func init() {
	flag.StringVar(&CLUSTER_URL, "cluster-url", "http://127.0.0.1:8080", "apiserver address to use with restclient")
}

func Logf(format string, a ...interface{}) {
	fmt.Fprintf(GinkgoWriter, format, a...)
}

func LoadConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags(CLUSTER_URL, "")
	Expect(err).NotTo(HaveOccurred())
	return config
}

func KubeClient() (*kubernetes.Clientset, error) {
	Logf("Using master %v\n", CLUSTER_URL)
	config := LoadConfig()
	clientset, err := kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	return clientset, nil
}

func GetAcClient() (client.Interface, error) {
	client, err := client.New(CLUSTER_URL)
	return client, err
}

func DeleteNS(clientset *kubernetes.Clientset, namespace *v1.Namespace) {
	defer GinkgoRecover()
	pods, err := clientset.Pods(namespace.Name).List(v1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	for _, pod := range pods.Items {
		clientset.Pods(namespace.Name).Delete(pod.Name, nil)
	}
	clientset.Namespaces().Delete(namespace.Name, nil)
}

func WaitForPod(clientset *kubernetes.Clientset, namespace string, name string, phase v1.PodPhase) *v1.Pod {
	defer GinkgoRecover()
	var podUpdated *v1.Pod
	Eventually(func() error {
		podUpdated, err := clientset.Core().Pods(namespace).Get(name)
		if err != nil {
			return err
		}
		if phase != "" && podUpdated.Status.Phase != phase {
			return fmt.Errorf("pod %v is not %v phase: %v", podUpdated.Name, phase, podUpdated.Status.Phase)
		}
		return nil
	}, 120*time.Second, 5*time.Second).Should(BeNil())
	return podUpdated
}

func DumpLogs(clientset *kubernetes.Clientset, pods ...v1.Pod) {
	for _, pod := range pods {
		dumpLogs(clientset, pod)
	}
}

func dumpLogs(clientset *kubernetes.Clientset, pod v1.Pod) {
	req := clientset.Core().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{})
	readCloser, err := req.Stream()
	Expect(err).NotTo(HaveOccurred())
	defer readCloser.Close()
	Logf("\n Dumping logs for %v:%v \n", pod.Namespace, pod.Name)
	_, err = io.Copy(GinkgoWriter, readCloser)
	Expect(err).NotTo(HaveOccurred())
}
