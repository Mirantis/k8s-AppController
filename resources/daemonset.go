package resources

import (
	"log"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/Mirantis/k8s-AppController/client"
	"github.com/Mirantis/k8s-AppController/interfaces"
)

//DaemonSet is wrapper for K8s DaemonSet object
type DaemonSet struct {
	DaemonSet *extensions.DaemonSet
	Client    unversioned.DaemonSetInterface
}

func daemonSetKey(name string) string {
	return "daemonset/" + name
}

func daemonSetStatus(d unversioned.DaemonSetInterface, name string) (string, error) {
	daemonSet, err := d.Get(name)
	if err != nil {
		return "error", err
	}
	if daemonSet.Status.CurrentNumberScheduled == daemonSet.Status.DesiredNumberScheduled {
		return "ready", nil
	}
	return "not ready", nil
}

//Key return DaemonSet key
func (d DaemonSet) Key() string {
	return daemonSetKey(d.DaemonSet.Name)
}

// Status returns DaemonSet status as a string "ready" means that its dependencies can be created
func (d DaemonSet) Status(meta map[string]string) (string, error) {
	return daemonSetStatus(d.Client, d.DaemonSet.Name)
}

//Create looks for DaemonSet in K8s and creates it if not present
func (d DaemonSet) Create() error {
	if err := checkExistence(d); err != nil {
		log.Println("Creating ", d.Key())
		d.DaemonSet, err = d.Client.Create(d.DaemonSet)
		return err
	}
	return nil
}

// Delete deletes DaemonSet from the cluster
func (d DaemonSet) Delete() error {
	return d.Client.Delete(d.DaemonSet.Name)
}

// NameMatches gets resource definition and a name and checks if
// the DaemonSet part of resource definition has matching name.
func (d DaemonSet) NameMatches(def client.ResourceDefinition, name string) bool {
	return def.DaemonSet != nil && def.DaemonSet.Name == name
}

// New returns new DaemonSet based on resource definition
func (d DaemonSet) New(def client.ResourceDefinition, c client.Interface) interfaces.Resource {
	return NewDaemonSet(def.DaemonSet, c.DaemonSets())
}

// NewExisting returns new ExistingDaemonSet based on resource definition
func (d DaemonSet) NewExisting(name string, c client.Interface) interfaces.Resource {
	return NewExistingDaemonSet(name, c.DaemonSets())
}

//NewDaemonSet is a constructor
func NewDaemonSet(daemonset *extensions.DaemonSet, client unversioned.DaemonSetInterface) DaemonSet {
	return DaemonSet{DaemonSet: daemonset, Client: client}
}

//ExistingDaemonSet is a wrapper for K8s DaemonSet object which is deployed on a cluster before AppController
type ExistingDaemonSet struct {
	Name   string
	Client unversioned.DaemonSetInterface
}

//Key returns DaemonSet name
func (d ExistingDaemonSet) Key() string {
	return daemonSetKey(d.Name)
}

// Status returns DaemonSet status as a string "ready" means that its dependencies can be created
func (d ExistingDaemonSet) Status(meta map[string]string) (string, error) {
	return daemonSetStatus(d.Client, d.Name)
}

//Create looks for existing DaemonSet and returns error if there is no such DaemonSet
func (d ExistingDaemonSet) Create() error {
	return createExistingResource(d)
}

// Delete deletes DaemonSet from the cluster
func (d ExistingDaemonSet) Delete() error {
	return d.Client.Delete(d.Name)
}

//NewExistingDaemonSet is a constructor
func NewExistingDaemonSet(name string, client unversioned.DaemonSetInterface) ExistingDaemonSet {
	return ExistingDaemonSet{Name: name, Client: client}
}
