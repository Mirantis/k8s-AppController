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

var _ = Describe("Basic Suite", func() {
	var clientset *kubernetes.Clientset
	var c client.Interface
	var framework GraphFramework
	var namespace *v1.Namespace

	BeforeEach(func() {
		By("Creating namespace and initializing test framework")
		var err error
		clientset, err = testutils.KubeClient()
		namespaceObj := &v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "e2e-tests-ac-",
				Namespace:    "",
			},
			Status: v1.NamespaceStatus{},
		}
		ns, err := clientset.Namespaces().Create(namespaceObj)
		Expect(err).NotTo(HaveOccurred())
		c, err = testutils.GetAcClient()
		Expect(err).NotTo(HaveOccurred())
		framework = GraphFramework{
			client:    c,
			clientset: clientset,
			ns:        ns.Name,
		}
		By("Deploying appcontroller image")
		framework.Prepare()
	})

	AfterEach(func() {
		By("Removing namespace")
		testutils.DeleteNS(clientset, namespace)
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
		parentPod := &v1.Pod{
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
		}
		childPod := PodPause("child-pod")
		framework.Connect(framework.Wrap(parentPod), framework.Wrap(childPod))
		framework.Run()
		testutils.WaitForPod(clientset, namespace.Name, parentPod.Name, "")
		time.Sleep(time.Second)
		_, err := clientset.Pods(namespace.Name).Get(childPod.Name)
		Expect(err).To(HaveOccurred())
	})

	It("Dependent Pod should be created if parent initialises correctly", func() {
		parentPod := PodPause("parent-pod")
		childPod := PodPause("child-pod")
		framework.Connect(framework.Wrap(parentPod), framework.Wrap(childPod))
		framework.Run()
		testutils.WaitForPod(clientset, namespace.Name, parentPod.Name, v1.PodRunning)
		testutils.WaitForPod(clientset, namespace.Name, childPod.Name, v1.PodRunning)
	})

	It("Service resdef works correctly with any k8s version", func() {
		pod1 := PodPause("pod1")
		pod1.Labels = map[string]string{"before": "service"}
		pod2 := PodPause("pod2")
		pod2.Labels = map[string]string{"after": "service"}
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
		svcWrapped := framework.Wrap(svc)
		By("Service depends on first pod")
		framework.Connect(framework.Wrap(pod1), svcWrapped)
		By("Second pod depends on service")
		framework.Connect(svcWrapped, framework.Wrap(pod2))
		framework.Run()
		By("Verifying that second pod will enter running state")
		testutils.WaitForPod(clientset, namespace.Name, pod2.Name, v1.PodRunning)
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
	client    client.Interface
	clientset *kubernetes.Clientset
	ns        string
}

func (g GraphFramework) Wrap(obj runtime.Object) *client.ResourceDefinition {
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

func (g GraphFramework) Connect(first, second *client.ResourceDefinition) {
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

func (g GraphFramework) Run() {
	cmd := exec.Command(
		"kubectl",
		"--namespace",
		g.ns,
		"exec",
		"k8s-appcontroller",
		"--",
		"ac-run",
		"-l",
		"ns:"+g.ns,
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

func (g GraphFramework) Prepare() {
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
							Value: g.ns,
						},
					},
				},
			},
		},
	}
	ac, err := g.clientset.Pods(g.ns).Create(appControllerObj)
	Expect(err).NotTo(HaveOccurred())
	testutils.WaitForPod(g.clientset, g.ns, ac.Name, v1.PodRunning)
	Eventually(func() bool {
		_, depsErr := g.client.ResourceDefinitions().List(api.ListOptions{})
		_, defsErr := g.client.Dependencies().List(api.ListOptions{})
		return defsErr == nil && depsErr == nil
	}, 120*time.Second, 5*time.Second).Should(BeTrue())
}

func PodPause(name string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
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
}
