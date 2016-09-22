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
	"io"
	"io/ioutil"
	"os"
	"testing"
)

// TestInput checks if input is properly retrieved from files
func TestInput(t *testing.T) {
	var inputTests = []struct {
		content  string
		indent   int
		expected string
	}{
		{"trololo\n lololo\n lololo\n\n", 1, " trololo\n  lololo\n  lololo\n \n"},
		{"trololo\n lololo\n lololo\n\n", 2, "  trololo\n   lololo\n   lololo\n  \n"},
	}

	for tc, tt := range inputTests {
		in, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		if err != nil {
			t.Fatal(err)
		}

		_, err = io.WriteString(in, tt.content)
		if err != nil {
			t.Fatal(err)
		}

		_, err = in.Seek(0, os.SEEK_SET)
		if err != nil {
			t.Fatal(err)
		}

		result := getInput(in, tt.indent)

		if result != tt.expected {
			t.Errorf("\"%s\" is not equal to \"%s\" in test case %d", result, tt.expected, tc+1)
		}
	}
}
