package integration

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"

	testutils "github.com/Mirantis/k8s-AppController/e2e/utils"
	"github.com/Mirantis/k8s-AppController/pkg/client"
)

type fixture struct {
	Namespace  *v1.Namespace
	Controller *v1.Pod
	ACClient   client.Interface
	Err        error
}

func runScheduler(clientset *kubernetes.Clientset, f fixture) {
	cmd := exec.Command(
		"kubectl",
		"--namespace",
		f.Namespace.Name,
		"exec",
		"k8s-appcontroller",
		"--",
		"ac-run",
		"-l",
		"ns:"+f.Namespace.Name,
	)
	out, err := cmd.Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			exErr := err.(*exec.ExitError)
			Fail(string(out) + string(exErr.Stderr))
		default:
			Expect(err).NotTo(HaveOccurred())

		}
	}
}

func getFixture(ch chan<- fixture, clientset *kubernetes.Clientset) {
	defer GinkgoRecover()
	namespaceObj := &v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "e2e-tests-ac-",
			Namespace:    "",
		},
		Status: v1.NamespaceStatus{},
	}
	ns, err := clientset.Namespaces().Create(namespaceObj)
	if err != nil {
		ch <- fixture{Err: err}
	}
	appControllerObj := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "k8s-appcontroller",
			Annotations: map[string]string{
				"pod.alpha.kubernetes.io/init-containers": `[{"name": "kubeac-bootstrap", "image": "mirantis/k8s-appcontroller", "imagePullPolicy": "Never", "command": ["kubeac", "bootstrap", "/opt/kubeac/manifests"]}]`,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Always",
			Containers: []v1.Container{
				{
					Name:            "kubeac",
					Image:           "mirantis/k8s-appcontroller",
					Command:         []string{"/usr/bin/run_runit"},
					ImagePullPolicy: v1.PullNever,
					Env: []v1.EnvVar{
						{
							Name:  "KUBERNETES_AC_LABEL_SELECTOR",
							Value: "",
						},
						{
							Name:  "KUBERNETES_AC_POD_NAMESPACE",
							Value: ns.ObjectMeta.Name,
						},
					},
				},
			},
		},
	}
	ac, err := clientset.Pods(ns.Name).Create(appControllerObj)
	if err != nil {
		ch <- fixture{Err: err}
	}
	testutils.WaitForPod(clientset, ns.Name, ac.Name, v1.PodRunning)
	client, err := testutils.GetAcClient()
	if err != nil {
		ch <- fixture{Err: err}
	}
	Eventually(func() bool {
		_, depsErr := client.ResourceDefinitions().List(api.ListOptions{})
		_, defsErr := client.Dependencies().List(api.ListOptions{})
		return defsErr == nil && depsErr == nil
	}, 120*time.Second, 5*time.Second).Should(BeTrue())
	ch <- fixture{Namespace: ns, Controller: ac, ACClient: client}
}

var _ = Describe("Basic Suite", func() {
	var clientset *kubernetes.Clientset
	var c client.Interface

	BeforeSuite(func() {
		var err error
		clientset, err = testutils.KubeClient()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("Removing all resource definitions")
		resDefs, err := c.ResourceDefinitions().List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, resDef := range resDefs.Items {
			err := c.ResourceDefinitions().Delete(resDef.Name, nil)
			Expect(err).NotTo(HaveOccurred())
		}
		By("Removing all dependencies")
		deps, err := c.Dependencies().List(api.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, dep := range deps.Items {
			err := c.Dependencies().Delete(dep.Name, nil)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Dependent Pod should not be created if parent is stuck in init", func() {
		ch := make(chan fixture)
		go getFixture(ch, clientset)
		f := <-ch
		Expect(f.Err).NotTo(HaveOccurred())
		c = f.ACClient
		defer testutils.DeleteNS(clientset, f.Namespace)
		parentPod := &client.ResourceDefinition{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "sleeper-parent",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Pod: &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "sleeper-parent",
					Annotations: map[string]string{
						"pod.alpha.kubernetes.io/init-containers": `[{"name": "sleeper-init", "image": "kubernetes/pause"}]`,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "sleeper",
							Image: "kubernetes/pause",
						},
					},
				},
			},
		}
		childPod := &client.ResourceDefinition{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "sleeper-child",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Pod: &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "sleeper-child",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "sleeper",
							Image: "kubernetes/pause",
						},
					},
				},
			},
		}
		parentPod, err := f.ACClient.ResourceDefinitions().Create(parentPod)
		Expect(err).NotTo(HaveOccurred())
		defer f.ACClient.ResourceDefinitions().Delete(parentPod.Name, nil)
		childPod, err = f.ACClient.ResourceDefinitions().Create(childPod)
		defer f.ACClient.ResourceDefinitions().Delete(childPod.Name, nil)
		Expect(err).NotTo(HaveOccurred())
		dep := &client.Dependency{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "dep",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Parent: "pod/sleeper-parent",
			Child:  "pod/sleeper-child",
			Meta:   map[string]string{},
		}
		dep, err = f.ACClient.Dependencies().Create(dep)
		Expect(err).NotTo(HaveOccurred())
		defer f.ACClient.Dependencies().Delete(dep.Name, nil)
		runScheduler(clientset, f)
		testutils.WaitForPod(clientset, f.Namespace.Name, "sleeper-parent", "")
		time.Sleep(time.Second)
		_, err = clientset.Pods(f.Namespace.Name).Get("sleeper-child")
		Expect(err).To(HaveOccurred())
	})

	It("Dependent Pod should be created if parent initialises correctly", func() {
		ch := make(chan fixture)
		go getFixture(ch, clientset)
		f := <-ch
		Expect(f.Err).NotTo(HaveOccurred())
		c = f.ACClient
		defer testutils.DeleteNS(clientset, f.Namespace)
		parentPod := &client.ResourceDefinition{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "dummy-parent",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Pod: &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "dummy-parent",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "sleeper",
							Image: "kubernetes/pause",
						},
					},
				},
			},
		}
		childPod := &client.ResourceDefinition{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "dummy-child",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Pod: &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "dummy-child",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "sleeper",
							Image: "kubernetes/pause",
						},
					},
				},
			},
		}
		parentPod, err := f.ACClient.ResourceDefinitions().Create(parentPod)
		Expect(err).NotTo(HaveOccurred())
		defer f.ACClient.ResourceDefinitions().Delete(parentPod.Name, nil)
		childPod, err = f.ACClient.ResourceDefinitions().Create(childPod)
		Expect(err).NotTo(HaveOccurred())
		defer f.ACClient.ResourceDefinitions().Delete(childPod.Name, nil)
		dep := &client.Dependency{
			ObjectMeta: api.ObjectMeta{
				GenerateName: "dep",
				Labels: map[string]string{
					"ns": f.Namespace.Name,
				},
			},
			Parent: "pod/dummy-parent",
			Child:  "pod/dummy-child",
			Meta:   map[string]string{},
		}
		dep, err = f.ACClient.Dependencies().Create(dep)
		defer f.ACClient.Dependencies().Delete(dep.Name, nil)
		Expect(err).NotTo(HaveOccurred())
		runScheduler(clientset, f)
		testutils.WaitForPod(clientset, f.Namespace.Name, "dummy-parent", "")
		time.Sleep(time.Second)
		testutils.WaitForPod(clientset, f.Namespace.Name, "dummy-child", v1.PodRunning)
	})

	It("Service resdef works correctly with any k8s version", func() {
		ch := make(chan fixture)
		go getFixture(ch, clientset)
		f := <-ch
		Expect(f.Err).NotTo(HaveOccurred())
		c = f.ACClient
		defer testutils.DeleteNS(clientset, f.Namespace)
		g := GraphFramework{client: f.ACClient, ns: f.Namespace.Name}
		pod1 := &v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:   "pod1",
				Labels: map[string]string{"before": "service"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "sleeper",
						Image: "kubernetes/pause",
					},
				},
			},
		}
		pod2 := &v1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:   "pod2",
				Labels: map[string]string{"after": "service"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "sleeper",
						Image: "kubernetes/pause",
					},
				},
			}}
		ports := []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 9999, TargetPort: intstr.FromInt(9999)}}
		svc := &v1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name: "svc1",
			},
			Spec: v1.ServiceSpec{
				Selector: pod1.Labels,
				Type:     v1.ServiceTypeNodePort,
				Ports:    ports,
			},
		}
		By("Creating two pods and one service")
		svcWrapped := g.wrap(svc)
		By("Service depends on first pod")
		g.connect(g.wrap(pod1), svcWrapped)
		By("Second pod depends on service")
		g.connect(svcWrapped, g.wrap(pod2))
		runScheduler(clientset, f)
		By("Verifying that second pod will enter running state")
		testutils.WaitForPod(clientset, f.Namespace.Name, pod2.Name, v1.PodRunning)
	})
})

func getKind(resdef *client.ResourceDefinition) string {
	if resdef.Pod != nil {
		return "pod"
	} else if resdef.Service != nil {
		return "service"
	}
	return ""
}

type GraphFramework struct {
	client client.Interface
	ns     string
}

func (g GraphFramework) wrap(obj runtime.Object) *client.ResourceDefinition {
	resdef := &client.ResourceDefinition{}
	switch v := obj.(type) {
	case *v1.Pod:
		resdef.Pod = v
		resdef.Name = v.Name
	case *v1.Service:
		resdef.Service = v
		resdef.Name = v.Name
	default:
		panic("Unknown type provided")
	}
	_, err := g.client.ResourceDefinitions().Create(resdef)
	Expect(err).NotTo(HaveOccurred())
	return resdef
}

func (g GraphFramework) connect(first, second *client.ResourceDefinition) {
	dep := &client.Dependency{
		ObjectMeta: api.ObjectMeta{
			GenerateName: "dep-",
			Labels: map[string]string{
				"ns": g.ns,
			},
		},
		Parent: getKind(first) + "/" + first.Name,
		Child:  getKind(second) + "/" + second.Name,
		Meta:   map[string]string{},
	}
	_, err := g.client.Dependencies().Create(dep)
	Expect(err).NotTo(HaveOccurred())
}
