package human

import (
	"strings"
)

// Indent indents every line
func Indent(indent int, data []string) []string {
	ret := make([]string, 0, cap(data))
	for _, line := range data {
		ret = append(ret, strings.Repeat(" ", indent)+line)
	}
	return ret
}
