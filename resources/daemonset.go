package resources

import (
	"errors"
	"log"

	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

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

func (d DaemonSet) UpdateMeta(meta map[string]string) error {
	return nil
}
func (d DaemonSet) Key() string {
	return daemonSetKey(d.DaemonSet.Name)
}

func (d DaemonSet) Status() (string, error) {
	return daemonSetStatus(d.Client, d.DaemonSet.Name)
}

func (d DaemonSet) Create() error {
	log.Println("Looking for daemonset", d.DaemonSet.Name)
	status, err := d.Status()

	if err == nil {
		log.Printf("Found daemonset %s, status: %s", d.DaemonSet.Name, status)
		log.Println("Skipping creation of daemonset", d.DaemonSet.Name)
	}
	log.Println("Creating daemonset", d.DaemonSet.Name)
	d.DaemonSet, err = d.Client.Create(d.DaemonSet)
	return err
}

func NewDaemonSet(daemonset *extensions.DaemonSet, client unversioned.DaemonSetInterface) DaemonSet {
	return DaemonSet{DaemonSet: daemonset, Client: client}
}

type ExistingDaemonSet struct {
	Name   string
	Client unversioned.DaemonSetInterface
}

func (d ExistingDaemonSet) UpdateMeta(meta map[string]string) error {
	return nil
}

func (d ExistingDaemonSet) Key() string {
	return daemonSetKey(d.Name)
}

func (d ExistingDaemonSet) Status() (string, error) {
	return daemonSetStatus(d.Client, d.Name)
}

func (d ExistingDaemonSet) Create() error {
	log.Println("Looking for daemonset", d.Name)
	status, err := d.Status()

	if err == nil {
		log.Printf("Found daemonset %s, status: %s", d.Name, status)
		return nil
	}

	log.Fatalf("DaemonSet %s not found", d.Name)
	return errors.New("DaemonSet not found")
}

func NewExistingDaemonSet(name string, client unversioned.DaemonSetInterface) ExistingDaemonSet {
	return ExistingDaemonSet{Name: name, Client: client}
}
