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

var (
	// ErrInvalidEnviron returned when environ has an incorrect format.
	ErrInvalidEnviron = errors.New("items in environ must have format key=value")
)

// EnvironToMap transforms a slice of string with the format "key=value" into
// the corresponding map of key-value pairs. If any item in environ does follow
// the format, EnvironToMap returns ErrInvalidEnviron.
func EnvironToMap(environ []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, v := range environ {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, ErrInvalidEnviron
		}
		m[parts[0]] = parts[1]
	}
	return m, nil
}

// MapToEnviron transforms a map of string to string into a slice of strings
// with the format "key=value".
func MapToEnviron(m map[string]string) []string {
	var environ []string
	for k, v := range m {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}
	return environ
}
