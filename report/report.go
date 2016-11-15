package report

// DependencyReport is a report of a single dependency of a node in graph
type DependencyReport struct {
	Dependency string
	Blocks     bool
	Percentage int
	Needed     int
	Message    string
}

// NodeReport is a report of a node in graph
type NodeReport struct {
	Dependent    string
	Blocked      bool
	Ready        bool
	Dependencies []DependencyReport
}

// DeploymentReport is a full report of the status of deployment
type DeploymentReport []NodeReport

// SimpleDependencyReport creates report for simple binary cases
func SimpleDependencyReport(name string, status string, err error) DependencyReport {
	if err != nil {
		return ErrorReport(name, err)
	}
	if status == "ready" {
		return DependencyReport{
			Dependency: name,
			Blocks:     false,
			Percentage: 100,
			Needed:     100,
			Message:    status,
		}
	}
	return DependencyReport{
		Dependency: name,
		Blocks:     true,
		Percentage: 0,
		Needed:     0,
		Message:    status,
	}
}

// ErrorReport creates a report for error cases
func ErrorReport(name string, err error) DependencyReport {
	return DependencyReport{
		Dependency: name,
		Blocks:     true,
		Percentage: 0,
		Needed:     100,
		Message:    err.Error(),
	}
}
