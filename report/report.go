package report

import (
	"github.com/Mirantis/k8s-AppController/interfaces"
)

// NodeReport is a report of a node in graph
type NodeReport struct {
	Dependent    string
	Blocked      bool
	Ready        bool
	Dependencies []interfaces.DependencyReport
}

// DeploymentReport is a full report of the status of deployment
type DeploymentReport []NodeReport

// SimpleReporter creates report for simple binary cases
type SimpleReporter struct {
	interfaces.Resource
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
func (r SimpleReporter) GetResource() interfaces.Resource {
	return r.Resource
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
