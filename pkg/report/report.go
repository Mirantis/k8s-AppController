package report

import (
	"fmt"
	"strings"

	"github.com/Mirantis/k8s-AppController/pkg/interfaces"
)

const ReportIndentSize = 4

// NodeReport is a report of a node in graph
type NodeReport struct {
	Dependent    string
	Blocked      bool
	Ready        bool
	Dependencies []interfaces.DependencyReport
}

// AsText returns a human-readable representation of the report as a slice
func (n NodeReport) AsText(indent int) []string {
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
		ret = append(ret, dependencyReportAsText(dependency, ReportIndentSize)...)
	}
	return Indent(indent, ret)
}

// DeploymentReport is a full report of the status of deployment
type DeploymentReport []NodeReport

// AsText returns a human-readable representation of the report as a slice
func (d DeploymentReport) AsText(indent int) []string {
	ret := make([]string, 0, len(d)*4)
	for _, n := range d {
		ret = append(ret, n.AsText(ReportIndentSize)...)
	}
	return Indent(indent, ret)
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
			Message:    string(status),
		}
	}
	return interfaces.DependencyReport{
		Dependency: r.Key(),
		Blocks:     true,
		Percentage: 0,
		Needed:     0,
		Message:    string(status),
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

// Indent indents every line
func Indent(indent int, data []string) []string {
	ret := make([]string, 0, cap(data))
	for _, line := range data {
		ret = append(ret, strings.Repeat(" ", indent)+line)
	}
	return ret
}

// dependencyReportAsText returns a human-readable representation of the report as a slice
func dependencyReportAsText(d interfaces.DependencyReport, indent int) []string {
	var blocksStr, percStr string
	if d.Blocks {
		blocksStr = "BLOCKS"
	} else {
		blocksStr = "DOESN'T BLOCK"
	}
	if d.Percentage == 100 {
		percStr = ""
	} else {
		percStr = fmt.Sprintf("%d%%/%d%%", d.Percentage, d.Needed)
	}
	ret := []string{
		fmt.Sprintf("Dependency: %s", d.Dependency),
		blocksStr,
	}
	if percStr != "" {
		ret = append(ret, percStr)
	}
	return Indent(indent, ret)
}
