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
	"strconv"
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

// TContext is a structure for CLI flags
type TContext struct {
	Examples string
	Version  string
}

// SkipIf14 makes test to be skipped when running on k8s 1.4
func SkipIf14() {
	if strings.Contains(TestContext.Version, "1.4") {
		Skip("This test is disabled on kubernetes of versions 1.4")
	}
}

// AddServiceAccountToAdmins will add system:serviceaccounts to cluster-admin ClusterRole
func AddServiceAccountToAdmins(c kubernetes.Interface) {
	if IsVersionOlderThan16() {
		return
	}
	By("Adding service account group to cluster-admin role")
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
	Expect(err).NotTo(HaveOccurred(), "Failed to create role binding for serviceaccounts")
}

// RemoveServiceAccountFromAdmins removes system:serviceaccounts from cluster-admin ClusterRole
func RemoveServiceAccountFromAdmins(c kubernetes.Interface) {
	if IsVersionOlderThan16() {
		return
	}
	By("Remowing service account group from cluster-admin role")
	err := c.Rbac().ClusterRoleBindings().Delete(rbacServiceAccountAdmin, nil)
	Expect(err).NotTo(HaveOccurred(), "Failed to remove serviceaccount from cluster-admin role")
}

// TestContext holds e2e CLI flags
var TestContext = TContext{}

var url string

func init() {
	flag.StringVar(&TestContext.Examples, "examples-directory", "examples", "Provide path to directory with examples")
	flag.StringVar(&TestContext.Version, "k8s-version", "", "Specify kubernetes version that is used for e2e tests")
	flag.StringVar(&url, "cluster-url", "http://127.0.0.1:8080", "apiserver address to use with restclient")
}

// SanitizedVersion parses K8s 2-component version string into float64 representation
func SanitizedVersion() float64 {
	var dirtyVersion string
	if strings.HasPrefix(TestContext.Version, "v") {
		dirtyVersion = TestContext.Version[1:]
	} else {
		dirtyVersion = TestContext.Version
	}
	version, err := strconv.ParseFloat(dirtyVersion, 64)
	Expect(err).NotTo(HaveOccurred())
	return version
}

// IsVersionOlderThan16 returns true, if k8s version is less than 1.6
func IsVersionOlderThan16() bool {
	return SanitizedVersion() < 1.6
}

// Logf prints formatted message to the tests log
func Logf(format string, a ...interface{}) {
	fmt.Fprintf(GinkgoWriter, format, a...)
}

// LoadConfig loads k8s client config
func LoadConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags(url, "")
	Expect(err).NotTo(HaveOccurred())
	return config
}

// KubeClient returns client to standard k8s entities
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

// DeleteNS deletes k8s namespace
func DeleteNS(clientset *kubernetes.Clientset, namespace *v1.Namespace) {
	defer GinkgoRecover()
	clientset.Namespaces().Delete(namespace.Name, nil)
}

// WaitForPod waits for k8s pod to get to specified running phase
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

// WaitForPodNotToBeCreated waits for pod to be created
func WaitForPodNotToBeCreated(clientset *kubernetes.Clientset, namespace string, name string) {
	defer GinkgoRecover()
	Consistently(func() bool {
		_, err := clientset.Core().Pods(namespace).Get(name)
		Expect(err).ToNot(BeNil())
		return errors.IsNotFound(err)
	}).Should(BeTrue())
}

// DumpLogs dumps pod logs
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

// ForEachYamlDocument executes action for each YAML document (parts of YAML file separated by "---")
func ForEachYamlDocument(data []byte, action func([]byte)) {
	matches := yamlDocumentDelimiter.FindAllIndex(data, -1)
	lastStart := 0
	for _, doc := range matches {
		action(data[lastStart:doc[0]])
		lastStart = doc[1]
	}
	action(data[lastStart:])
}
