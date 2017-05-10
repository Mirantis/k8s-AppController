// Copyright 2017 Mirantis
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

// Based on the work of Joel Scoble:
// https://github.com/mohae/deepcopy/blob/master/deepcopy.go
// (The MIT License)

package copier

import (
	"reflect"
	"regexp"
	"time"
)

var varsRe = regexp.MustCompile("\\$\\w+\\b")

// Copy creates a deep copy of whatever is passed to it and returns the copy
// in an interface{}.  The returned value will need to be asserted to the
// correct type.
func Copy(src interface{}) interface{} {
	return CopyWithReplacements(src, func(p string) string {
		return p
	})
}

// CopyWithReplacements does deep copy of the object and performs string substitution for the string fields
func CopyWithReplacements(src interface{}, replacementsFunc func(string) string, replaceIn ...string) interface{} {
	if src == nil {
		return nil
	}

	// Make the interface a reflect.Value
	original := reflect.ValueOf(src)

	// Make a copy of the same type as the original.
	cpy := reflect.New(original.Type()).Elem()

	_, found := addToPath("", "", replaceIn...)
	// Recursively copy the original.
	if found {
		copyRecursive(original, cpy, replacementsFunc, "", "*")
	} else {
		copyRecursive(original, cpy, replacementsFunc, "", replaceIn...)
	}

	// Return the copy as an interface.
	return cpy.Interface()
}

// copyRecursive does the actual copying of the interface. It currently has
// limited support for what it can handle. Add as needed.
func copyRecursive(original, cpy reflect.Value, replacementsFunc func(string) string, path string, replaceIn ...string) {
	// handle according to original's Kind
	switch original.Kind() {
	case reflect.Ptr:
		// Get the actual value being pointed to.
		originalValue := original.Elem()

		// if  it isn't valid, return.
		if !originalValue.IsValid() {
			return
		}
		cpy.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, cpy.Elem(), replacementsFunc, path, replaceIn...)

	case reflect.Interface:
		// If this is a nil, don't do anything
		if original.IsNil() {
			return
		}
		// Get the value for the interface, not the pointer.
		originalValue := original.Elem()

		// Get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue, replacementsFunc, path, replaceIn...)
		cpy.Set(copyValue)

	case reflect.Struct:
		t, ok := original.Interface().(time.Time)
		if ok {
			cpy.Set(reflect.ValueOf(t))
			return
		}
		// Go through each field of the struct and copy it.
		for i := 0; i < original.NumField(); i++ {
			// The Type's StructField for a given field is checked to see if StructField.PkgPath
			// is set to determine if the field is exported or not because CanSet() returns false
			// for settable fields.  I'm not sure why.  -mohae
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			name := original.Type().Field(i).Name
			newPath, found := addToPath(path, name, replaceIn...)
			if found {
				copyRecursive(original.Field(i), cpy.Field(i), replacementsFunc, newPath, "*")
			} else {
				copyRecursive(original.Field(i), cpy.Field(i), replacementsFunc, newPath, replaceIn...)
			}
		}

	case reflect.Slice:
		if original.IsNil() {
			return
		}
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i), replacementsFunc, path, replaceIn...)
		}

	case reflect.Map:
		if original.IsNil() {
			return
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			value := original.MapIndex(key)
			copyValue := reflect.New(value.Type()).Elem()
			newPath, found := addToPath(path, "Values", replaceIn...)
			if found {
				copyRecursive(value, copyValue, replacementsFunc, newPath, "*")
			} else {
				copyRecursive(value, copyValue, replacementsFunc, newPath, replaceIn...)
			}
			newPath, found = addToPath(path, "Keys", replaceIn...)
			copyKey := reflect.New(key.Type()).Elem()
			if found {
				copyRecursive(key, copyKey, replacementsFunc, newPath, "*")
			} else {
				copyRecursive(key, copyKey, replacementsFunc, newPath, replaceIn...)
			}

			cpy.SetMapIndex(copyKey, copyValue)
		}
	case reflect.String:
		for _, r := range replaceIn {
			if path == r || r == "*" {
				cpy.Set(reflect.ValueOf(EvaluateString(original.String(), replacementsFunc)).Convert(original.Type()))
				return
			}
		}
		cpy.Set(original)
	default:
		cpy.Set(original)
	}
}

func addToPath(path, name string, replaceIn ...string) (string, bool) {
	var newPath string
	if path == "" {
		newPath = name
	} else {
		newPath = path + "." + name
	}
	for _, r := range replaceIn {
		if r == newPath || r == "*" {
			return newPath, true
		}
	}
	return newPath, false
}

func EvaluateString(template string, replacementsFunc func(string) string) string {
	return varsRe.ReplaceAllStringFunc(template, func(p string) string {
		return replacementsFunc(p[1:])
	})
}
