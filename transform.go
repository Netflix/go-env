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
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidEnviron returned when environ has an incorrect format.
var ErrInvalidEnviron = errors.New("items in environ must have format key=value")

// EnvSet represents a set of environment variables.
type EnvSet map[string]string

// ChangeSet represents a set of environment variables changes, corresponding to
// os.Setenv and os.Unsetenv operations.
type ChangeSet map[string]*string

// Apply applies a ChangeSet to EnvSet, modifying its contents.
func (e EnvSet) Apply(cs ChangeSet) {
	for k, v := range cs {
		if v == nil {
			// Equivalent to os.Unsetenv
			delete(e, k)
			continue
		}
		// Equivalent to os.Setenv
		e[k] = *v
	}
}

// EnvironToEnvSet transforms a slice of string with the format "key=value" into
// the corresponding EnvSet. If any item in environ does follow the format,
// EnvironToEnvSet returns ErrInvalidEnviron.
func EnvironToEnvSet(environ []string) (EnvSet, error) {
	// We error out the function on the first invalid item found, so we can
	// optimistically pre-allocate the EnvSet map with the correct size and
	// let the GC clean up in the invalid/exit case alongside the function call.
	m := make(EnvSet, len(environ))
	for _, v := range environ {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, ErrInvalidEnviron
		}
		m[parts[0]] = parts[1]
	}
	return m, nil
}

// EnvSetToEnviron transforms a EnvSet into a slice of strings with the format
// "key=value".
func EnvSetToEnviron(m EnvSet) []string {
	environ := make([]string, 0, len(m))
	for k, v := range m {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}
	return environ
}
