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

// Package env provides an `env` struct field tag to marshal and unmarshal
// environment variables.
package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

var (
	// ErrInvalidValue returned when the value passed to Unmarshal is nil or not a
	// pointer to a struct.
	ErrInvalidValue = errors.New("value must be a non-nil pointer to a struct")

	// ErrUnsupportedType returned when a field with tag "env" is unsupported.
	ErrUnsupportedType = errors.New("field is an unsupported type")

	// ErrUnexportedField returned when a field with tag "env" is not exported.
	ErrUnexportedField = errors.New("field must be exported")
)

// Unmarshal parses an environment mapping and stores the result in the value
// pointed to by v. If v is nil or not a pointer to a struct, Unmarshal returns
// an ErrInvalidValue.
//
// Fields tagged with "env" will have the unmarshalled data of the matching key
// from data. If the tagged field is not exported, Unmarshal returns
// ErrUnexportedField.
//
// If the field has a type that is unsupported, Unmarshal returns
// ErrUnsupportedType.
func Unmarshal(data map[string]string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrInvalidValue
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fieldValue := rv.Field(i)
		switch fieldValue.Kind() {
		case reflect.Struct:
			if !fieldValue.Addr().CanInterface() {
				continue
			}

			iface := fieldValue.Addr().Interface()
			err := Unmarshal(data, iface)
			if err != nil {
				return err
			}
		}

		structField := t.Field(i)
		tag := structField.Tag.Get("env")
		if tag == "" {
			continue
		}

		if !fieldValue.CanSet() {
			return ErrUnexportedField
		}

		envVar, ok := data[tag]
		if !ok {
			continue
		}

		err := set(fieldValue, envVar)
		if err != nil {
			return err
		}
	}

	return nil
}

func set(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int:
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(v))
	default:
		return ErrUnsupportedType
	}

	return nil
}

// UnmarshalFromEnviron parses an environment mapping from os.Environ and
// stores the result in the value pointed to by v. If v is nil or not a
// pointer to a struct, UnmarshalFromEnviron returns an ErrInvalidValue.
//
// Fields tagged with "env" will have the unmarshalled data of the matching key
// from data. If the tagged field is not exported, UnmarshalFromEnviron returns
// ErrUnexportedField.
//
// If the field has a type that is unsupported, UnmarshalFromEnviron returns
// ErrUnsupportedType.
func UnmarshalFromEnviron(v interface{}) error {
	m, err := EnvironToMap(os.Environ())
	if err != nil {
		return err
	}

	return Unmarshal(m, v)
}

// Marshal returns an environment mapping of v. If v is nil or not a pointer,
// Marshal returns an ErrInvalidValue.
//
// Marshal uses fmt.Sprintf to transform encountered values to its default
// string format. Values without the "env" field tag are ignored.
//
// Nested structs are traversed recursively.
func Marshal(v interface{}) (map[string]string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil, ErrInvalidValue
	}

	data := make(map[string]string)
	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fieldValue := rv.Field(i)
		switch fieldValue.Kind() {
		case reflect.Struct:
			if !fieldValue.Addr().CanInterface() {
				continue
			}

			iface := fieldValue.Addr().Interface()
			nestedData, err := Marshal(iface)
			if err != nil {
				return nil, err
			}

			for k, v := range nestedData {
				data[k] = v
			}
		}

		structField := t.Field(i)
		tag := structField.Tag.Get("env")
		if tag == "" {
			continue
		}

		data[tag] = fmt.Sprintf("%v", fieldValue.Interface())
	}

	return data, nil
}
