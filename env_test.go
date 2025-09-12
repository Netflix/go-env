// Copyright 2018 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package env

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

type ValidStruct struct {
	// Home should match Environ because it has a "env" field tag.
	Home string `env:"HOME"`

	// Jenkins should be recursed into.
	Jenkins struct {
		Workspace string `env:"WORKSPACE"`

		// PointerMissing should not be set if the environment variable is missing.
		PointerMissing *string `env:"JENKINS_POINTER_MISSING"`
	}

	// PointerString should be nil if unset, with "" being a valid value.
	PointerString *string `env:"POINTER_STRING"`

	// PointerInt should work along with other supported types.
	PointerInt *int `env:"POINTER_INT"`

	// PointerUint should work along with other supported types.
	PointerUint *uint `env:"POINTER_UINT"`

	// PointerPointerString should be recursed into.
	PointerPointerString **string `env:"POINTER_POINTER_STRING"`

	// PointerMissing should not be set if the environment variable is missing.
	PointerMissing *string `env:"POINTER_MISSING"`

	// Extra should remain with a zero-value because it has no "env" field tag.
	Extra string

	// Additional supported types
	Int     int     `env:"INT"`
	Uint    uint    `env:"UINT"`
	Float32 float32 `env:"FLOAT32"`
	Float64 float64 `env:"FLOAT64"`
	Bool    bool    `env:"BOOL"`

	MultipleTags string `env:"npm_config_cache,NPM_CONFIG_CACHE"`

	MultipleTagsWithDefault string `env:"multiple_tags_with_default,MULTIPLE_TAGS_WITH_DEFAULT,default=default_tags_value"`

	TagWithDefault string `env:"tag_with_default,default=default_tag_value"`

	TagWithRequired string `env:"tag_with_required,required=false"`

	TagWithSeparator string `env:"tag_with_separator,separator=&"`

	// time.Duration is supported
	Duration time.Duration `env:"TYPE_DURATION"`

	// Custom unmarshaler should support scalar types
	Base64EncodedString Base64EncodedString `env:"BASE64_ENCODED_STRING"`
	// Custom unmarshaler should support struct types
	JSONData JSONData `env:"JSON_DATA"`
	// Custom unmarshaler should support pointer types as well
	PointerJSONData *JSONData `env:"POINTER_JSON_DATA"`
}

type UnsupportedStruct struct {
	Timestamp time.Time `env:"TIMESTAMP"`
}

type UnexportedStruct struct {
	home string `env:"HOME"`
}

type DefaultValueStruct struct {
	DefaultString             string        `env:"MISSING_STRING,default=found"`
	DefaultKeyValueString     string        `env:"MISSING_KVSTRING,default=key=value"`
	DefaultBool               bool          `env:"MISSING_BOOL,default=true"`
	DefaultInt                int           `env:"MISSING_INT,default=7"`
	DefaultUint               uint          `env:"MISSING_UINT,default=4294967295"`
	DefaultFloat32            float32       `env:"MISSING_FLOAT32,default=8.9"`
	DefaultFloat64            float64       `env:"MISSING_FLOAT64,default=10.11"`
	DefaultDuration           time.Duration `env:"MISSING_DURATION,default=5s"`
	DefaultStringSlice        []string      `env:"MISSING_STRING_SLICE,default=separate|values"`
	DefaultSliceWithSeparator []string      `env:"ANOTHER_MISSING_STRING_SLICE,default=separate&values,separator=&"`
	DefaultRequiredSlice      []string      `env:"MISSING_REQUIRED_DEFAULT,default=other|things,required=true"`
	DefaultWithOptionsMissing string        `env:"MISSING_1,MISSING_2,default=present"`
	DefaultWithOptionsPresent string        `env:"MISSING_1,PRESENT,default=present"`
}

type RequiredValueStruct struct {
	Required            string `env:"REQUIRED_VAL,required=true"`
	RequiredMore        string `env:"REQUIRED_VAL_MORE,required=true"`
	RequiredWithDefault string `env:"REQUIRED_MISSING,default=myValue,required=true"`
	NotRequired         string `env:"NOT_REQUIRED,required=false"`
	InvalidExtra        string `env:"INVALID,invalid=invalid"`
}

type Base64EncodedString string

func (b *Base64EncodedString) UnmarshalEnvironmentValue(data string) error {
	value, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	*b = Base64EncodedString(value)
	return nil
}

func (b Base64EncodedString) MarshalEnvironmentValue() (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(b)), nil
}

type JSONData struct {
	SomeField int `json:"someField"`
}

func (j *JSONData) UnmarshalEnvironmentValue(data string) error {
	var tmp JSONData
	if err := json.Unmarshal([]byte(data), &tmp); err != nil {
		return err
	}
	*j = tmp
	return nil
}

func (j JSONData) MarshalEnvironmentValue() (string, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type IterValuesStruct struct {
	StringSlice   []string        `env:"STRING"`
	IntSlice      []int           `env:"INT"`
	Int64Slice    []int64         `env:"INT64"`
	DurationSlice []time.Duration `env:"DURATION"`
	BoolSlice     []bool          `env:"BOOL"`
	KVStringSlice []string        `env:"KV"`
	WithSeparator []int           `env:"SEPARATOR,separator=&"`
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{
			"HOME":             "/home/test",
			"WORKSPACE":        "/mnt/builds/slave/workspace/test",
			"EXTRA":            "extra",
			"INT":              "1",
			"UINT":             "4294967295",
			"FLOAT32":          "2.3",
			"FLOAT64":          "4.5",
			"BOOL":             "true",
			"npm_config_cache": "first",
			"NPM_CONFIG_CACHE": "second",
			"TYPE_DURATION":    "5s",
		}
		validStruct ValidStruct
	)

	if err := Unmarshal(environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != "/home/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/home/test", validStruct.Home)
	}

	if validStruct.Jenkins.Workspace != "/mnt/builds/slave/workspace/test" {
		t.Errorf(
			"Expected field value to be '%s' but got '%s'",
			"/mnt/builds/slave/workspace/test",
			validStruct.Jenkins.Workspace,
		)
	}

	if validStruct.PointerString != nil {
		t.Errorf("Expected field value to be '%v' but got '%v'", nil, validStruct.PointerString)
	}

	if validStruct.Extra != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", validStruct.Extra)
	}

	if validStruct.Int != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, validStruct.Int)
	}

	if validStruct.Uint != 4294967295 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 4294967295, validStruct.Uint)
	}

	if validStruct.Float32 != 2.3 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 2.3, validStruct.Float32)
	}

	if validStruct.Float64 != 4.5 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 4.5, validStruct.Float64)
	}

	if validStruct.Bool != true {
		t.Errorf("Expected field value to be '%t' but got '%t'", true, validStruct.Bool)
	}

	if validStruct.MultipleTags != "first" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "first", validStruct.MultipleTags)
	}

	if validStruct.Duration != 5*time.Second {
		t.Errorf("Expected field value to be '%s' but got '%s'", "5s", validStruct.Duration)
	}

	if v, ok := environ["HOME"]; ok {
		t.Errorf("Expected field '%s' to not exist but got '%s'", "HOME", v)
	}

	if v, ok := environ["EXTRA"]; !ok {
		t.Errorf("Expected field '%s' to exist but missing", "EXTRA")
	} else if v != "extra" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "extra", v)
	}
}

func TestUnmarshalPointer(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{
			"POINTER_STRING":         "",
			"POINTER_INT":            "1",
			"POINTER_UINT":           "4294967295",
			"POINTER_POINTER_STRING": "",
		}
		validStruct ValidStruct
	)

	if err := Unmarshal(environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.PointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else if *validStruct.PointerString != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", *validStruct.PointerString)
	}

	if validStruct.PointerInt == nil {
		t.Errorf("Expected field value to be '%d' but got '%v'", 1, nil)
	} else if *validStruct.PointerInt != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, *validStruct.PointerInt)
	}

	if validStruct.PointerUint == nil {
		t.Errorf("Expected field value to be '%d' but got '%v'", 4294967295, nil)
	} else if *validStruct.PointerUint != 4294967295 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 4294967295, *validStruct.PointerUint)
	}

	if validStruct.PointerPointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else {
		if *validStruct.PointerPointerString == nil {
			t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
		} else if **validStruct.PointerPointerString != "" {
			t.Errorf("Expected field value to be '%s' but got '%s'", "", **validStruct.PointerPointerString)
		}
	}

	if validStruct.PointerMissing != nil {
		t.Errorf("Expected field value to be '%v' but got '%s'", nil, *validStruct.PointerMissing)
	}
}

func TestCustomUnmarshal(t *testing.T) {
	t.Parallel()
	var (
		someValue = "some value"
		environ   = map[string]string{
			"BASE64_ENCODED_STRING": base64.StdEncoding.EncodeToString([]byte(someValue)),
			"JSON_DATA":             `{ "someField": 42 }`,
			"POINTER_JSON_DATA":     `{ "someField": 43 }`,
		}
		validStruct ValidStruct
	)

	if err := Unmarshal(environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Base64EncodedString != Base64EncodedString(someValue) {
		t.Errorf("Expected field value to be '%s' but got '%s'", someValue, string(validStruct.Base64EncodedString))
	}

	if validStruct.PointerJSONData == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", someValue, nil)
	} else if validStruct.PointerJSONData.SomeField != 43 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 43, validStruct.PointerJSONData.SomeField)
	}

	if validStruct.JSONData.SomeField != 42 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 42, validStruct.JSONData.SomeField)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	t.Parallel()
	var (
		environ     = make(map[string]string)
		validStruct ValidStruct
	)

	if err := Unmarshal(environ, validStruct); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}

	ptr := &validStruct
	if err := Unmarshal(environ, &ptr); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}
}

func TestUnmarshalUnsupported(t *testing.T) {
	t.Parallel()
	var (
		environ           = map[string]string{"TIMESTAMP": "2016-07-15T12:00:00.000Z"}
		unsupportedStruct UnsupportedStruct
	)

	if err := Unmarshal(environ, &unsupportedStruct); !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("Expected error 'ErrUnsupportedType' but got '%s'", err)
	}
}

func TestUnmarshalFromEnviron(t *testing.T) {
	t.Parallel()
	environ := os.Environ()

	es, err := EnvironToEnvSet(environ)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	home := es["HOME"]

	var validStruct ValidStruct
	es, err = UnmarshalFromEnviron(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != home {
		t.Errorf("Expected environment variable to be '%s' but got '%s'", home, validStruct.Home)
	}

	if v, ok := es["HOME"]; ok {
		t.Errorf("Expected field '%s' to not exist but got '%s'", "HOME", v)
	}
}

func TestUnmarshalUnexported(t *testing.T) {
	t.Parallel()
	var (
		environ          = map[string]string{"HOME": "/home/edgarl"}
		unexportedStruct UnexportedStruct
	)

	if err := Unmarshal(environ, &unexportedStruct); !errors.Is(err, ErrUnexportedField) {
		t.Errorf("Expected error 'ErrUnexportedField' but got '%s'", err)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{
			"STRING":    "separate|values",
			"INT":       "1|2",
			"INT64":     "3|4",
			"DURATION":  "60s|70h",
			"BOOL":      "true|false",
			"KV":        "k1=v1|k2=v2",
			"SEPARATOR": "1&2", // struct has `separator=&`
		}
		iterValStruct IterValuesStruct
	)

	if err := Unmarshal(environ, &iterValStruct); err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}

	testCases := [][]interface{}{
		{iterValStruct.StringSlice, []string{"separate", "values"}},
		{iterValStruct.IntSlice, []int{1, 2}},
		{iterValStruct.Int64Slice, []int64{3, 4}},
		{iterValStruct.DurationSlice, []time.Duration{time.Second * 60, time.Hour * 70}},
		{iterValStruct.BoolSlice, []bool{true, false}},
		{iterValStruct.KVStringSlice, []string{"k1=v1", "k2=v2"}},
		{iterValStruct.WithSeparator, []int{1, 2}},
	}
	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase[0], testCase[1]) {
			t.Errorf("Expected field value to be '%v' but got '%v'", testCase[1], testCase[0])
		}
	}
}

func TestUnmarshalDefaultValues(t *testing.T) {
	t.Parallel()
	var (
		environ            = map[string]string{"PRESENT": "youFoundMe"}
		defaultValueStruct DefaultValueStruct
	)

	if err := Unmarshal(environ, &defaultValueStruct); err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	testCases := [][]interface{}{
		{defaultValueStruct.DefaultInt, 7},
		{defaultValueStruct.DefaultUint, uint(4294967295)},
		{defaultValueStruct.DefaultFloat32, float32(8.9)},
		{defaultValueStruct.DefaultFloat64, 10.11},
		{defaultValueStruct.DefaultBool, true},
		{defaultValueStruct.DefaultString, "found"},
		{defaultValueStruct.DefaultKeyValueString, "key=value"},
		{defaultValueStruct.DefaultDuration, 5 * time.Second},
		{defaultValueStruct.DefaultStringSlice, []string{"separate", "values"}},
		{defaultValueStruct.DefaultSliceWithSeparator, []string{"separate", "values"}},
		{defaultValueStruct.DefaultRequiredSlice, []string{"other", "things"}},
		{defaultValueStruct.DefaultWithOptionsMissing, "present"},
		{defaultValueStruct.DefaultWithOptionsPresent, "youFoundMe"},
	}
	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase[0], testCase[1]) {
			t.Errorf("Expected field value to be '%v' but got '%v'", testCase[1], testCase[0])
		}
	}
}

func TestUnmarshalRequiredValues(t *testing.T) {
	t.Parallel()
	var requiredValuesStruct RequiredValueStruct

	// Try missing REQUIRED_VAL and REQUIRED_VAL_MORE
	environ := make(map[string]string)
	err := Unmarshal(environ, &requiredValuesStruct)
	if err == nil {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}
	errMissing := ErrMissingRequiredValue{Value: "REQUIRED_VAL"}
	if err.Error() != errMissing.Error() {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}

	// Fill REQUIRED_VAL and retry REQUIRED_VAL_MORE
	environ = make(map[string]string)
	environ["REQUIRED_VAL"] = "required"
	err = Unmarshal(environ, &requiredValuesStruct)
	if err == nil {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}
	errMissing = ErrMissingRequiredValue{Value: "REQUIRED_VAL_MORE"}
	if err.Error() != errMissing.Error() {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}

	// Fill both required values
	environ = make(map[string]string)
	environ["REQUIRED_VAL"] = "required"
	environ["REQUIRED_VAL_MORE"] = "required"
	if err = Unmarshal(environ, &requiredValuesStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}
	if requiredValuesStruct.Required != "required" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "required", requiredValuesStruct.Required)
	}
	if requiredValuesStruct.RequiredWithDefault != "myValue" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "myValue", requiredValuesStruct.RequiredWithDefault)
	}
}

func TestMarshal(t *testing.T) {
	t.Parallel()
	validStruct := ValidStruct{
		Home: "/home/test",
		Jenkins: struct {
			Workspace      string  `env:"WORKSPACE"`
			PointerMissing *string `env:"JENKINS_POINTER_MISSING"`
		}{
			Workspace: "/mnt/builds/slave/workspace/test",
		},
		Extra:                   "extra",
		Int:                     1,
		Uint:                    4294967295,
		Float32:                 float32(2.3),
		Float64:                 4.5,
		Bool:                    true,
		MultipleTags:            "foobar",
		MultipleTagsWithDefault: "baz",
		TagWithDefault:          "bar",
		TagWithRequired:         "foo",
		TagWithSeparator:        "val1&val2",
		Duration:                3 * time.Minute,
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

	if environ["UINT"] != "4294967295" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "2", environ["UINT"])
	}

	if environ["FLOAT32"] != "2.3" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "2.3", environ["FLOAT32"])
	}

	if environ["FLOAT64"] != "4.5" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "4.5", environ["FLOAT64"])
	}

	if environ["BOOL"] != "true" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "true", environ["BOOL"])
	}

	if environ["npm_config_cache"] != "foobar" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "foobar", environ["npm_config_cache"])
	}

	if environ["NPM_CONFIG_CACHE"] != "foobar" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "foobar", environ["NPM_CONFIG_CACHE"])
	}

	if environ["multiple_tags_with_default"] != "baz" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "baz", environ["multiple_tags_with_default"])
	}

	if environ["default=default_tags_value"] != "" {
		t.Errorf("'default=default_tags_value' not expected to be a valid field value.")
	}

	if environ["tag_with_default"] != "bar" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "bar", environ["tag_with_default"])
	}

	if environ["tag_with_required"] != "foo" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "foo", environ["tag_with_required"])
	}

	if environ["tag_with_separator"] != "val1&val2" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "val1&val2", environ["tag_with_separator"])
	}

	if environ["required=true"] != "" {
		t.Errorf("'required=true' not expected to be a valid field value.")
	}

	if environ["separator=&"] != "" {
		t.Errorf("'separator=&' not expected to be a valid field value.")
	}

	if environ["default=default_tag_value"] != "" {
		t.Errorf("'default=default_tag_value' not expected to be a valid field value.")
	}

	if environ["TYPE_DURATION"] != "3m0s" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "3m0s", environ["TYPE_DURATION"])
	}
}

func TestMarshalInvalid(t *testing.T) {
	t.Parallel()
	var validStruct ValidStruct
	if _, err := Marshal(validStruct); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}

	ptr := &validStruct
	if _, err := Marshal(&ptr); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}
}

func TestMarshalPointer(t *testing.T) {
	t.Parallel()
	var (
		empty       = ""
		validStruct = ValidStruct{PointerString: &empty}
	)

	es, err := Marshal(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if v, ok := es["POINTER_STRING"]; !ok {
		t.Errorf("Expected field '%s' to exist but missing", "POINTER_STRING")
	} else if v != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", v)
	}

	if v, ok := es["POINTER_MISSING"]; ok {
		t.Errorf("Expected field '%s' to not exist but got '%s'", "POINTER_MISSING", v)
	}

	if v, ok := es["JENKINS_POINTER_MISSING"]; ok {
		t.Errorf("Expected field '%s' to not exist but got '%s'", "JENKINS_POINTER_MISSING", v)
	}
}

func TestMarshalCustom(t *testing.T) {
	t.Parallel()
	var (
		someValue   = Base64EncodedString("someValue")
		validStruct = ValidStruct{
			Base64EncodedString: someValue,
			JSONData:            JSONData{SomeField: 42},
			PointerJSONData:     &JSONData{SomeField: 43},
		}
	)

	es, err := Marshal(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if v, ok := es["BASE64_ENCODED_STRING"]; !ok {
		t.Errorf("Expected field '%s' to exist but missing", "BASE64_ENCODED_STRING")
	} else if v != base64.StdEncoding.EncodeToString([]byte(someValue)) {
		t.Errorf("Expected field value to be '%s' but got '%s'", base64.StdEncoding.EncodeToString([]byte(someValue)), v)
	}

	if v, ok := es["JSON_DATA"]; !ok {
		t.Errorf("Expected field '%s' to exist but got '%s'", "JSON_DATA", v)
	} else if v != `{"someField":42}` {
		t.Errorf("Expected field value to be '%s' but got '%s'", `{"someField":42}`, v)
	}

	if v, ok := es["POINTER_JSON_DATA"]; !ok {
		t.Errorf("Expected field '%s' to exist but got '%s'", "POINTER_JSON_DATA", v)
	} else if v != `{"someField":43}` {
		t.Errorf("Expected field value to be '%s' but got '%s'", `{"someField":43}`, v)
	}
}

// Config represents the example configuration from the user's request
type Config struct {
	Server   ServerConfig   `env:"prefix=SERVER_"`
	Database DatabaseConfig `env:"prefix=DATABASE_"`
	JWT      JWTConfig
	Email    EmailConfig
	Redis    RedisConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `env:"HOST,default=localhost"`
	Port string `env:"PORT,default=8080"`
	Mode string `env:"MODE,default=debug"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `env:"HOST,default=localhost"`
	Port     string `env:"PORT,default=5432"`
	Username string `env:"USERNAME,default=postgres"`
	Password string `env:"PASSWORD"`
	Database string `env:"DATABASE,default=myapp"`
}

// JWTConfig holds JWT configuration (no prefix)
type JWTConfig struct {
	Secret     string `env:"JWT_SECRET"`
	ExpiryTime string `env:"JWT_EXPIRY,default=24h"`
}

// EmailConfig holds email configuration (no prefix)
type EmailConfig struct {
	SMTPHost string `env:"EMAIL_SMTP_HOST,default=localhost"`
	SMTPPort string `env:"EMAIL_SMTP_PORT,default=587"`
}

// RedisConfig holds Redis configuration (no prefix)
type RedisConfig struct {
	Host string `env:"REDIS_HOST,default=localhost"`
	Port string `env:"REDIS_PORT,default=6379"`
}

func TestUnmarshalWithPrefix(t *testing.T) {
	envs := map[string]string{
		"SERVER_HOST":       "api.example.com",
		"SERVER_PORT":       "3000",
		"SERVER_MODE":       "production",
		"DATABASE_HOST":     "db.example.com",
		"DATABASE_PORT":     "5433",
		"DATABASE_USERNAME": "admin",
		"DATABASE_PASSWORD": "secret123",
		"DATABASE_DATABASE": "production_db",
		"JWT_SECRET":        "super-secret-key",
		"JWT_EXPIRY":        "48h",
		"EMAIL_SMTP_HOST":   "smtp.example.com",
		"EMAIL_SMTP_PORT":   "465",
		"REDIS_HOST":        "cache.example.com",
		"REDIS_PORT":        "6380",
	}

	var config Config
	err := Unmarshal(envs, &config)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	// Check Server config (should use SERVER_ prefix)
	if config.Server.Host != "api.example.com" {
		t.Errorf("Expected Server.Host to be 'api.example.com' but got '%s'", config.Server.Host)
	}
	if config.Server.Port != "3000" {
		t.Errorf("Expected Server.Port to be '3000' but got '%s'", config.Server.Port)
	}
	if config.Server.Mode != "production" {
		t.Errorf("Expected Server.Mode to be 'production' but got '%s'", config.Server.Mode)
	}

	// Check Database config (should use DATABASE_ prefix)
	if config.Database.Host != "db.example.com" {
		t.Errorf("Expected Database.Host to be 'db.example.com' but got '%s'", config.Database.Host)
	}
	if config.Database.Port != "5433" {
		t.Errorf("Expected Database.Port to be '5433' but got '%s'", config.Database.Port)
	}
	if config.Database.Username != "admin" {
		t.Errorf("Expected Database.Username to be 'admin' but got '%s'", config.Database.Username)
	}
	if config.Database.Password != "secret123" {
		t.Errorf("Expected Database.Password to be 'secret123' but got '%s'", config.Database.Password)
	}
	if config.Database.Database != "production_db" {
		t.Errorf("Expected Database.Database to be 'production_db' but got '%s'", config.Database.Database)
	}

	// Check JWT config (no prefix)
	if config.JWT.Secret != "super-secret-key" {
		t.Errorf("Expected JWT.Secret to be 'super-secret-key' but got '%s'", config.JWT.Secret)
	}
	if config.JWT.ExpiryTime != "48h" {
		t.Errorf("Expected JWT.ExpiryTime to be '48h' but got '%s'", config.JWT.ExpiryTime)
	}

	// Check Email config (no prefix)
	if config.Email.SMTPHost != "smtp.example.com" {
		t.Errorf("Expected Email.SMTPHost to be 'smtp.example.com' but got '%s'", config.Email.SMTPHost)
	}
	if config.Email.SMTPPort != "465" {
		t.Errorf("Expected Email.SMTPPort to be '465' but got '%s'", config.Email.SMTPPort)
	}

	// Check Redis config (no prefix)
	if config.Redis.Host != "cache.example.com" {
		t.Errorf("Expected Redis.Host to be 'cache.example.com' but got '%s'", config.Redis.Host)
	}
	if config.Redis.Port != "6380" {
		t.Errorf("Expected Redis.Port to be '6380' but got '%s'", config.Redis.Port)
	}
}

func TestUnmarshalWithPrefixDefaults(t *testing.T) {
	// Test that defaults work correctly with prefixes when environment variables are missing
	envs := map[string]string{
		"SERVER_HOST": "custom.example.com", // Only set one server value
		"JWT_SECRET":  "jwt-key",            // Set JWT secret but not expiry
	}

	var config Config
	err := Unmarshal(envs, &config)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	// Check Server config (should use defaults for missing values)
	if config.Server.Host != "custom.example.com" {
		t.Errorf("Expected Server.Host to be 'custom.example.com' but got '%s'", config.Server.Host)
	}
	if config.Server.Port != "8080" { // Should use default
		t.Errorf("Expected Server.Port to be '8080' (default) but got '%s'", config.Server.Port)
	}
	if config.Server.Mode != "debug" { // Should use default
		t.Errorf("Expected Server.Mode to be 'debug' (default) but got '%s'", config.Server.Mode)
	}

	// Check Database config (should use all defaults)
	if config.Database.Host != "localhost" {
		t.Errorf("Expected Database.Host to be 'localhost' (default) but got '%s'", config.Database.Host)
	}
	if config.Database.Port != "5432" {
		t.Errorf("Expected Database.Port to be '5432' (default) but got '%s'", config.Database.Port)
	}

	// Check JWT config
	if config.JWT.Secret != "jwt-key" {
		t.Errorf("Expected JWT.Secret to be 'jwt-key' but got '%s'", config.JWT.Secret)
	}
	if config.JWT.ExpiryTime != "24h" { // Should use default
		t.Errorf("Expected JWT.ExpiryTime to be '24h' (default) but got '%s'", config.JWT.ExpiryTime)
	}
}

func TestMarshalWithPrefix(t *testing.T) {
	config := Config{
		Server: ServerConfig{
			Host: "api.example.com",
			Port: "3000",
			Mode: "production",
		},
		Database: DatabaseConfig{
			Host:     "db.example.com",
			Port:     "5433",
			Username: "admin",
			Password: "secret123",
			Database: "production_db",
		},
		JWT: JWTConfig{
			Secret:     "super-secret-key",
			ExpiryTime: "48h",
		},
		Email: EmailConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: "465",
		},
		Redis: RedisConfig{
			Host: "cache.example.com",
			Port: "6380",
		},
	}

	es, err := Marshal(&config)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	// Check Server config (should have SERVER_ prefix)
	if v, ok := es["SERVER_HOST"]; !ok || v != "api.example.com" {
		t.Errorf("Expected SERVER_HOST to be 'api.example.com' but got '%s' (exists: %v)", v, ok)
	}
	if v, ok := es["SERVER_PORT"]; !ok || v != "3000" {
		t.Errorf("Expected SERVER_PORT to be '3000' but got '%s' (exists: %v)", v, ok)
	}
	if v, ok := es["SERVER_MODE"]; !ok || v != "production" {
		t.Errorf("Expected SERVER_MODE to be 'production' but got '%s' (exists: %v)", v, ok)
	}

	// Check Database config (should have DATABASE_ prefix)
	if v, ok := es["DATABASE_HOST"]; !ok || v != "db.example.com" {
		t.Errorf("Expected DATABASE_HOST to be 'db.example.com' but got '%s' (exists: %v)", v, ok)
	}
	if v, ok := es["DATABASE_PORT"]; !ok || v != "5433" {
		t.Errorf("Expected DATABASE_PORT to be '5433' but got '%s' (exists: %v)", v, ok)
	}

	// Check JWT config (no prefix)
	if v, ok := es["JWT_SECRET"]; !ok || v != "super-secret-key" {
		t.Errorf("Expected JWT_SECRET to be 'super-secret-key' but got '%s' (exists: %v)", v, ok)
	}
	if v, ok := es["JWT_EXPIRY"]; !ok || v != "48h" {
		t.Errorf("Expected JWT_EXPIRY to be '48h' but got '%s' (exists: %v)", v, ok)
	}

	// Check that non-prefixed vars don't have prefix
	if _, ok := es["HOST"]; ok {
		t.Errorf("Expected HOST (without prefix) to not exist in EnvSet")
	}
	if _, ok := es["PORT"]; ok {
		t.Errorf("Expected PORT (without prefix) to not exist in EnvSet")
	}
}

func TestNestedPrefixes(t *testing.T) {
	// Test nested structs with multiple levels of prefixes
	type Level2 struct {
		Value string `env:"VALUE"`
	}
	type Level1 struct {
		Nested Level2 `env:"prefix=LEVEL2_"`
		Direct string `env:"DIRECT"`
	}
	type Root struct {
		Level1 Level1 `env:"prefix=LEVEL1_"`
	}

	envs := map[string]string{
		"LEVEL1_LEVEL2_VALUE": "nested_value",
		"LEVEL1_DIRECT":       "direct_value",
	}

	var root Root
	err := Unmarshal(envs, &root)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if root.Level1.Nested.Value != "nested_value" {
		t.Errorf("Expected Level1.Nested.Value to be 'nested_value' but got '%s'", root.Level1.Nested.Value)
	}
	if root.Level1.Direct != "direct_value" {
		t.Errorf("Expected Level1.Direct to be 'direct_value' but got '%s'", root.Level1.Direct)
	}
}
