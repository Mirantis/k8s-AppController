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
	"regexp"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/rbac/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var yamlDocumentDelimiter = regexp.MustCompile("(?m)^---")
const rbacServiceAccountAdmin = "system:serviceaccount-admin"

type TContext struct {
	Examples string
	Version  string
}

func SkipIf14() {
	// TODO semver compare
	if strings.Contains(TestContext.Version, "1.4") {
		Skip("This test is disabled on kubernetes of version 1.4")
	}
}

// AddServiceAccountToAdmins will add system:serviceaccounts to cluster-admin ClusterRole
func AddServiceAccountToAdmins(c kubernetes.Interface) {
	if strings.Contains(TestContext.Version, "1.6") {
		return
	}
	roleBinding := &v1alpha1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: rbacServiceAccountAdmin,
		},
		Subjects: []v1alpha1.Subject{{
			Kind: "Group",
			Name: "system:serviceaccounts",
		}},
		RoleRef: v1alpha1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := c.Rbac().ClusterRoleBindings().Create(roleBinding)
	Expect(err).NotTo(HaveOccurred(), "Wasnt able to create role binding for serviceaccounts")
}

func RemoveServiceAccountFromAdmins(c kubernetes.Interface) {
	err := c.Rbac().ClusterRoleBindings().Delete(rbacServiceAccountAdmin, nil)
	Expect(err).NotTo(HaveOccurred(), "Failed to remove serviceaccount from admin group")
}

var TestContext = TContext{}

var url string

func init() {
	flag.StringVar(&TestContext.Examples, "examples-directory", "examples", "Provide path to directory with examples")
	flag.StringVar(&TestContext.Version, "k8s-version", "", "Specify kubernetes version that is used for e2e tests")
	flag.StringVar(&url, "cluster-url", "http://127.0.0.1:8080", "apiserver address to use with restclient")
}

func Logf(format string, a ...interface{}) {
	fmt.Fprintf(GinkgoWriter, format, a...)
}

func LoadConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags(url, "")
	Expect(err).NotTo(HaveOccurred())
	return config
}

func KubeClient() (*kubernetes.Clientset, error) {
	Logf("Using master %v\n", url)
	config := LoadConfig()
	clientset, err := kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	return clientset, nil
}

// GetAcClient returns client for given namespace which will be used in e2e tests
func GetAcClient(namespace string) (client.Interface, error) {
	client, err := client.NewForNamespace(url, namespace)
	return client, err
}

func DeleteNS(clientset *kubernetes.Clientset, namespace *v1.Namespace) {
	defer GinkgoRecover()
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
	}).Should(BeNil())
	return podUpdated
}

func WaitForPodNotToBeCreated(clientset *kubernetes.Clientset, namespace string, name string) {
	defer GinkgoRecover()
	Consistently(func() bool {
		_, err := clientset.Core().Pods(namespace).Get(name)
		Expect(err).ToNot(BeNil())
		return errors.IsNotFound(err)
	}).Should(BeTrue())
}

func DumpLogs(clientset *kubernetes.Clientset, pods ...*v1.Pod) {
	for _, pod := range pods {
		dumpLogs(clientset, pod)
	}
}

func dumpLogs(clientset *kubernetes.Clientset, pod *v1.Pod) {
	req := clientset.Core().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{})
	readCloser, err := req.Stream()
	Expect(err).NotTo(HaveOccurred())
	defer readCloser.Close()
	Logf("\n Dumping logs for %v:%v \n", pod.Namespace, pod.Name)
	_, err = io.Copy(GinkgoWriter, readCloser)
	Expect(err).NotTo(HaveOccurred())
}

func ForEachYamlDocument(data []byte, action func([]byte)) {
	matches := yamlDocumentDelimiter.FindAllIndex(data, -1)
	lastStart := 0
	for _, doc := range matches {
		action(data[lastStart: doc[0]])
		lastStart = doc[1]
	}
	action(data[lastStart:])
}
