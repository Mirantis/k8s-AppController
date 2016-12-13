package human

import (
	"github.com/fatih/color"
	"strings"
)

// MaybeColorFactory returns functions that color the string optionally if
// a boolean flag is set
func MaybeColorFactory(colorAt color.Attribute) func(string, bool) string {
	colorObj := color.New(colorAt)
	return func(s string, useColor bool) string {
		if useColor {
			color.Red(s)
			colorObj.Println(s)
			return colorObj.SprintFunc()(s)
		} else {
			return s
		}
	}
}

var Red = MaybeColorFactory(color.FgRed)
var Green = MaybeColorFactory(color.FgGreen)
var Yellow = MaybeColorFactory(color.FgYellow)

// Indent indents every line
func Indent(indent int, data []string) []string {
	ret := make([]string, 0, cap(data))
	for _, line := range data {
		ret = append(ret, strings.Repeat(" ", indent)+line)
	}
	return ret
}
