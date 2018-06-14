/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parameters

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var keymapRegex = regexp.MustCompile(`^([^\[]+)\[(.+)\]\s*$`)

// ParseVariableJSON converts a JSON object into a map of keys and values
// Example:
// `{ "location" : "east", "group" : "demo" }' becomes map[location:east group:demo]
func ParseVariableJSON(params string) (map[string]interface{}, error) {
	var p map[string]interface{}
	err := json.Unmarshal([]byte(params), &p)
	if err != nil {
		return nil, fmt.Errorf("invalid parameters (%s)", params)
	}
	return p, nil
}

// ParseVariableAssignments converts a string array of variable assignments
// into a map of keys and values
// Example:
// [a=b c=abc1232=== d=banana d=pineapple] becomes map[a:b c:abc1232=== d:[banana pineapple]]
func ParseVariableAssignments(params []string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for _, p := range params {

		parts := strings.SplitN(p, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in name=value format", p)
		}

		variable := strings.TrimSpace(parts[0])
		if variable == "" {
			return nil, fmt.Errorf("invalid parameter (%s), variable name is required", p)
		}
		value := strings.TrimSpace(parts[1])

		storedValue, ok := variables[variable]
		// Logic to add new value to map variables:
		// if variable DNE: add pair to variables as variable:value
		// if variable exists in form of variable:value, create array to hold old value&new value
		// if variable exists in form variable:[some values], append new value to existing array
		if !ok {
			variables[variable] = value // if there is no key, add key&value as string
		} else {
			switch storedValType := storedValue.(type) {
			case string:
				variables[variable] = []string{storedValType, value}
			case []string:
				variables[variable] = append(storedValType, value)
			}
		}
	}

	return variables, nil
}

// ParseKeyMaps converts a string array of key lookups
// into a map of the map name and key
// Example:
// [a[b] mysecret[foo.txt]] becomes map[a:b mysecret:foo.txt]
func ParseKeyMaps(params []string) (map[string]string, error) {
	keymap := map[string]string{}

	for _, p := range params {
		parts := keymapRegex.FindStringSubmatch(p)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in MAP[KEY] format", p)
		}

		mapName := strings.TrimSpace(parts[1])
		if mapName == "" {
			return nil, fmt.Errorf("invalid parameter (%s), map is required", p)
		}

		key := strings.TrimSpace(parts[2])
		if key == "" {
			return nil, fmt.Errorf("invalid parameter (%s), key is required", p)
		}

		keymap[mapName] = key
	}

	return keymap, nil
}
