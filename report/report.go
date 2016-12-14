package report

import (
	"fmt"

	"github.com/Mirantis/k8s-AppController/human"
	"github.com/Mirantis/k8s-AppController/interfaces"
)

// NodeReport is a report of a node in graph
type NodeReport struct {
	Dependent    string
	Blocked      bool
	Ready        bool
	Dependencies []interfaces.DependencyReport
}

// AsHuman returns a human-readable representation of the report as a slice
func (n NodeReport) AsHuman(indent int) []string {
	var blockedStr, readyStr string
	if n.Blocked {
		blockedStr = "BLOCKED"
	} else {
		blockedStr = "NOT BLOCKED"
	}

	if n.Ready {
		readyStr = "READY"
	} else {
		readyStr = "NOT READY"
	}

	ret := []string{
		fmt.Sprintf("Resource: %s", n.Dependent),
		blockedStr,
		readyStr,
	}
	for _, dependency := range n.Dependencies {
		ret = append(ret, dependency.AsHuman(4)...)
	}
	return human.Indent(indent, ret)
}

// DeploymentReport is a full report of the status of deployment
type DeploymentReport []NodeReport

// AsHuman returns a human-readable representation of the report as a slice
func (d DeploymentReport) AsHuman(indent int) []string {
	ret := make([]string, 0, len(d)*4)
	for _, n := range d {
		ret = append(ret, n.AsHuman(4)...)
	}
	return human.Indent(indent, ret)
}

// SimpleReporter creates report for simple binary cases
type SimpleReporter struct {
	interfaces.BaseResource
}

// GetDependencyReport returns a dependency report for this reporter
func (r SimpleReporter) GetDependencyReport(meta map[string]string) interfaces.DependencyReport {
	status, err := r.Status(meta)
	if err != nil {
		return ErrorReport(r.Key(), err)
	}
	if status == "ready" {
		return interfaces.DependencyReport{
			Dependency: r.Key(),
			Blocks:     false,
			Percentage: 100,
			Needed:     100,
			Message:    status,
		}
	}
	return interfaces.DependencyReport{
		Dependency: r.Key(),
		Blocks:     true,
		Percentage: 0,
		Needed:     0,
		Message:    status,
	}
}

// GetResource returns the underlying resource
func (r SimpleReporter) GetResource() interfaces.BaseResource {
	return r.BaseResource
}

// ErrorReport creates a report for error cases
func ErrorReport(name string, err error) interfaces.DependencyReport {
	return interfaces.DependencyReport{
		Dependency: name,
		Blocks:     true,
		Percentage: 0,
		Needed:     100,
		Message:    err.Error(),
	}
}
