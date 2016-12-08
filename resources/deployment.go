package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
	"github.com/Mirantis/k8s-AppController/report"
)

// Deployment is wrapper for K8s Deployment object
type Deployment struct {
	Base
	Deployment *extensions.Deployment
	Client     unversioned.DeploymentInterface
}

func deploymentKey(name string) string {
	return "deployment/" + name
}

func deploymentStatus(d unversioned.DeploymentInterface, name string) (string, error) {
	deployment, err := d.Get(name)
	if err != nil {
		return "error", err
	}

	if deployment.Status.UpdatedReplicas >= deployment.Spec.Replicas && deployment.Status.AvailableReplicas >= deployment.Spec.Replicas {
		return "ready", nil
	}
	return "not ready", nil
}

// Key return Deployment key
func (d Deployment) Key() string {
	return deploymentKey(d.Deployment.Name)
}

// Status returns Deployment status as a string "ready" means that its dependencies can be created
func (d Deployment) Status(meta map[string]string) (string, error) {
	return deploymentStatus(d.Client, d.Deployment.Name)
}

// Create looks for Deployment in K8s and creates it if not present
func (d Deployment) Create() error {
	log.Println("Looking for deployment", d.Deployment.Name)
	status, err := d.Status(nil)

	if err == nil {
		log.Printf("Found deployment %s, status: %s", d.Deployment.Name, status)
		log.Println("Skipping creation of deployment", d.Deployment.Name)
	}
	log.Println("Creating deployment", d.Deployment.Name)
	d.Deployment, err = d.Client.Create(d.Deployment)
	return err
}

// Delete deletes Deployment from the cluster
func (d Deployment) Delete() error {
	return d.Client.Delete(d.Deployment.Name, nil)
}

// NameMatches gets resource definition and a name and checks if
// the Deployment part of resource definition has matching name.
func (d Deployment) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.Deployment != nil && def.Deployment.Name == name
}

// New returns new Deployment based on resource definition
func (d Deployment) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewDeployment(def.Deployment, c.Deployments(), def.Meta)
}

// NewExisting returns new ExistingDeployment based on resource definition
func (d Deployment) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingDeployment(name, c.Deployments())
}

// NewDeployment is a constructor
func NewDeployment(deployment *extensions.Deployment, client unversioned.DeploymentInterface, meta map[string]string) interfaces.Resource {
	return report.SimpleReporter{BaseResource: Deployment{Base: Base{meta}, Deployment: deployment, Client: client}}
}

// ExistingDeployment is a wrapper for K8s Deployment object which is deployed on a cluster before AppController
type ExistingDeployment struct {
	Base
	Name   string
	Client unversioned.DeploymentInterface
}

// UpdateMeta does nothing at the moment
func (d ExistingDeployment) UpdateMeta(meta map[string]string) error {
	return nil
}

// Key returns Deployment name
func (d ExistingDeployment) Key() string {
	return deploymentKey(d.Name)
}

// Status returns Deployment status as a string "ready" means that its dependencies can be created
func (d ExistingDeployment) Status(meta map[string]string) (string, error) {
	return deploymentStatus(d.Client, d.Name)
}

// Create looks for existing Deployment and returns error if there is no such Deployment
func (d ExistingDeployment) Create() error {
	log.Println("Looking for deployment", d.Name)
	status, err := d.Status(nil)

	if err == nil {
		log.Printf("Found deployment %s, status: %s", d.Name, status)
		return nil
	}

	log.Fatalf("Deployment %s not found", d.Name)
	return errors.New("Deployment not found")
}

// Delete deletes Deployment from the cluster
func (d ExistingDeployment) Delete() error {
	return d.Client.Delete(d.Name, nil)
}

// NewExistingDeployment is a constructor
func NewExistingDeployment(name string, client unversioned.DeploymentInterface) interfaces.Resource {
	return report.SimpleReporter{BaseResource: ExistingDeployment{Name: name, Client: client}}
}
