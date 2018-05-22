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
	"os"
	"testing"
	"time"
)

type ValidStruct struct {
	// Home should match Environ because it has a "env" field tag.
	Home string `env:"HOME"`

	// Jenkins should be recursed into.
	Jenkins struct {
		Workspace string `env:"WORKSPACE"`
	}

	// Extra should remain with a zero-value because it has no "env" field tag.
	Extra string

	// Additional supported types
	Int  int  `env:"INT"`
	Bool bool `env:"BOOL"`
}

type UnsupportedStruct struct {
	Timestamp time.Time `env:"TIMESTAMP"`
}

type UnexportedStruct struct {
	home string `env:"HOME"`
}

func TestUnmarshal(t *testing.T) {
	environ := map[string]string{
		"HOME":      "/home/test",
		"WORKSPACE": "/mnt/builds/slave/workspace/test",
		"EXTRA":     "extra",
		"INT":       "1",
		"BOOL":      "true",
	}

	var validStruct ValidStruct
	err := Unmarshal(environ, &validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != environ["HOME"] {
		t.Errorf("Expected field value to be '%s' but got '%s'", environ["HOME"], validStruct.Home)
	}

	if validStruct.Jenkins.Workspace != environ["WORKSPACE"] {
		t.Errorf("Expected field value to be '%s' but got '%s'", environ["WORKSPACE"], validStruct.Jenkins.Workspace)
	}

	if validStruct.Extra != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", validStruct.Extra)
	}

	if validStruct.Int != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, validStruct.Int)
	}

	if validStruct.Bool != true {
		t.Errorf("Expected field value to be '%t' but got '%t'", true, validStruct.Bool)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	environ := make(map[string]string)

	var validStruct ValidStruct
	err := Unmarshal(environ, validStruct)
	if err != ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}

	ptr := &validStruct
	err = Unmarshal(environ, &ptr)
	if err != ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}
}

func TestUnmarshalUnsupported(t *testing.T) {
	environ := map[string]string{
		"TIMESTAMP": "2016-07-15T12:00:00.000Z",
	}

	var unsupportedStruct UnsupportedStruct
	err := Unmarshal(environ, &unsupportedStruct)
	if err != ErrUnsupportedType {
		t.Errorf("Expected error 'ErrUnsupportedType' but got '%s'", err)
	}
}

func TestUnmarshalFromEnviron(t *testing.T) {
	environ := os.Environ()

	m, err := EnvironToMap(environ)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	var validStruct ValidStruct
	err = UnmarshalFromEnviron(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != m["HOME"] {
		t.Errorf("Expected environment variable to be '%s' but got '%s'", m["HOME"], validStruct.Home)
	}
}

func TestUnmarshalUnexported(t *testing.T) {
	environ := map[string]string{
		"HOME": "/home/edgarl",
	}

	var unexportedStruct UnexportedStruct
	err := Unmarshal(environ, &unexportedStruct)
	if err != ErrUnexportedField {
		t.Errorf("Expected error 'ErrUnexportedField' but got '%s'", err)
	}
}

func TestMarshal(t *testing.T) {
	validStruct := ValidStruct{
		Home: "/home/test",
		Jenkins: struct {
			Workspace string `env:"WORKSPACE"`
		}{
			Workspace: "/mnt/builds/slave/workspace/test",
		},
		Extra: "extra",
		Int:   1,
		Bool:  true,
	}

	environ, err := Marshal(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if environ["HOME"] != validStruct.Home {
		t.Errorf("Expected field value to be '%s' but got '%s'", environ["HOME"], environ["HOME"])
	}

	if environ["WORKSPACE"] != validStruct.Jenkins.Workspace {
		t.Errorf("Expected field value to be '%s' but got '%s'", environ["WORKSPACE"], environ["WORKSPACE"])
	}

	if environ["EXTRA"] != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", environ["EXTRA"])
	}

	if environ["INT"] != "1" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "1", environ["INT"])
	}

	if environ["BOOL"] != "true" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "true", environ["BOOL"])
	}
}

func TestMarshalInvalid(t *testing.T) {
	var validStruct ValidStruct
	_, err := Marshal(validStruct)
	if err != ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}

	ptr := &validStruct
	_, err = Marshal(&ptr)
	if err != ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}
}
