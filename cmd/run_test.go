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
	"os"
	"testing"
)

// TestEmptyLabel checks if label is empty if no values are provided
func TestEmptyLabel(t *testing.T) {
	cmd, _ := InitRunCommand()
	label, _ := getLabelSelector(cmd)

	if label != "" {
		t.Errorf("label selector should be empty, is `%s` instead", label)
	}
}

// TestLabelEnv checks if label selector is retrieved from env variable
func TestLabelEnv(t *testing.T) {
	cmd, _ := InitRunCommand()
	val := "TEST_KEY=TEST_VALUE"
	os.Setenv("KUBERNETES_AC_LABEL_SELECTOR", val)
	label, _ := getLabelSelector(cmd)

	if label != val {
		t.Errorf("label selector should be equal to `%s`, is `%s` instead", val, label)
	}
}

// TestLabelFlag checks if label selector is retrieved from command flag and if it overwrites env var
func TestLabelFlag(t *testing.T) {
	cmd, _ := InitRunCommand()

	val := "TEST_KEY=TEST_VALUE"
	val2 := "TEST_OTHER_KEY=TEST_OTHER_VALUE"
	os.Setenv("KUBERNETES_AC_LABEL_SELECTOR", val)
	cmd.Flags().Parse([]string{"-l", val2})

	label, _ := getLabelSelector(cmd)

	if label != val2 {
		t.Errorf("label selector should be equal to `%s`, is `%s` instead", val2, label)
	}
}
