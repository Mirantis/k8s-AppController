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
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// RootCmd is top-level AppController command. It is not executable, but it has sub-commands attached
var RootCmd *cobra.Command

// Init initializes RootCmd, adds flags to subcommands and attaches subcommands to root command
func Init() {
	var err error
	var labelSelector string
	Run.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector. Overrides APPCONTROLLER_LABEL_SELECTOR env variable in AppController pod.")

	concurrencyString := os.Getenv("KUBERNETES_AC_CONCURRENCY")

	var concurrencyDefault int
	if len(concurrencyString) > 0 {
		concurrencyDefault, err = strconv.Atoi(concurrencyString)
		if err != nil {
			log.Printf("KUBERNETES_AC_CONCURRENCY is set to '%s' but it does not look like an integer: %v",
				concurrencyString, err)
			concurrencyDefault = 0
		}
	}
	var concurrency int
	Run.Flags().IntVarP(&concurrency, "concurrency", "c", concurrencyDefault, "concurrency")

	var format string
	Wrap.Flags().StringVarP(&format, "format", "f", "yaml", "file format")

	RootCmd = &cobra.Command{Use: "ac"}
	RootCmd.AddCommand(Bootstrap, Run, Wrap)
}
