// Copyright 2018 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package env

import (
	"fmt"
	"testing"
)

func TestEnvironToMap(t *testing.T) {
	environ := []string{"HOME=/home/edgarl", "WORKSPACE=/mnt/builds/slave/workspace/test"}

	m, err := EnvironToMap(environ)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if m["HOME"] != "/home/edgarl" {
		t.Errorf("Expected map value to be '%s' but got '%s'", "/home/edgarl", m["HOME"])
	}

	if m["WORKSPACE"] != "/mnt/builds/slave/workspace/test" {
		t.Errorf("Expected map value to be '%s' but got '%s'", "/mnt/builds/slave/workspace/test", m["WORKSPACE"])
	}
}

func TestEnvironToMapInvalid(t *testing.T) {
	environ := []string{"INVALID"}

	_, err := EnvironToMap(environ)
	if err != ErrInvalidEnviron {
		t.Errorf("Expected 'ErrInvalidEnviron' but got '%s'", err)
	}
}

func TestMapToEnviron(t *testing.T) {
	m := map[string]string{
		"HOME":      "/home/test",
		"WORKSPACE": "/mnt/builds/slave/workspace/test",
	}

	environ := MapToEnviron(m)
	if len(environ) != 2 {
		t.Errorf("Expected environ to have %d items but instead got %d", 2, len(environ))
	}

	for k, v := range m {
		found := false
		envPair := fmt.Sprintf("%s=%s", k, v)
		for _, e := range environ {
			if e == envPair {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected environ to contain '%s'", envPair)
		}
	}
}
