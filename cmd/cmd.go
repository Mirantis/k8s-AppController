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
	"log"

	"github.com/spf13/cobra"
)

// RootCmd is top-level AppController command. It is not executable, but it has sub-commands attached
var RootCmd *cobra.Command

// Init initializes RootCmd, adds flags to subcommands and attaches subcommands to root command
func Init() {
	run, err := InitRunCommand()
	if err != nil {
		log.Fatal(err)
	}
	status, err := InitGetStatusCommand()
	if err != nil {
		log.Fatal(err)
	}

	var format string
	Wrap.Flags().StringVarP(&format, "format", "f", "yaml", "file format")

	RootCmd = &cobra.Command{Use: "kubeac"}
	RootCmd.AddCommand(Bootstrap, run, Wrap, status, Deploy)
}
