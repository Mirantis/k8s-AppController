package resources

import (
	"testing"

	"github.com/Mirantis/k8s-AppController/mocks"
)

//TestDaemonSetSuccessCheck check status for ready DaemonSet
func TestDaemonSetSuccessCheck(t *testing.T) {
	c := mocks.NewClient()
	status, err := daemonSetStatus(c.DaemonSets(), "not-fail")

	if err != nil {
		t.Error(err)
	}
	if status != "ready" {
		t.Errorf("Status should be ready , is %s instead", status)
	}
}

//TestDaemonSetFailCheck status of not ready daemonset
func TestDaemonSetFailCheck(t *testing.T) {
	c := mocks.NewClient()
	status, err := daemonSetStatus(c.DaemonSets(), "fail")
	if err != nil {
		t.Error(err)
	}
	if status != "not ready" {
		t.Errorf("Status should be not ready, is %s instead.", status)
	}
}
