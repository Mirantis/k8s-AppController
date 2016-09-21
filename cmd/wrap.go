// Copyright 2016 Mirantis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Mirantis/k8s-AppController/cmd/format"
)

func getInput(stream *os.File, indent int) string {
	result := ""
	spaces := strings.Repeat(" ", indent)

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		// add spaces for identation
		result += spaces + scanner.Text() + "\n"
	}
	return result
}

func wrap(cmd *cobra.Command, args []string) {
	fileFormat, err := cmd.Flags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	var f format.Format
	switch fileFormat {
	case "yaml":
		f = format.Yaml{}
	case "json":
		f = format.Json{}
	default:
		log.Fatal("Unknonwn file format. Expected one of: yaml, json")
	}

	definition := getInput(os.Stdin, f.IndentLevel())

	out, err := f.Wrap(definition)
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}

var Wrap = &cobra.Command{
	Use:   "wrap",
	Short: "Echo wrapped k8s object to stdout",
	Long:  "Echo wrapped k8s object to stdout",
	Run:   wrap,
}
