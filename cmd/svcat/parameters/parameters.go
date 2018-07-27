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
// into a map of keys and values.
// Examples:
// [a=b c=abc1232=== d=X d=Y e.f.g=Z] --> map[a:b c:abc1232=== d:[X Y] e:map[f:map[g:Z]]]
func ParseVariableAssignments(params []string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	for _, p := range params {
		var dotKeys []string

		parts := strings.SplitN(p, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in name=value format", p)
		}

		variable := strings.TrimSpace(parts[0])
		if variable == "" {
			return nil, fmt.Errorf("invalid parameter (%s), variable name is required", p)
		}
		value := strings.TrimSpace(parts[1])

		if strings.ContainsAny(variable, ".") {
			//check if dot params form is correct: xxx.xx
			r, _ := regexp.Compile(`[^\s\.]+\.[^\s\.]+(\.[^\s\.]+)*`)
			if r.MatchString(variable) == true {
				dotKeys = strings.Split(variable, ".")
				variable = dotKeys[0]
			} else {
				return nil, fmt.Errorf("invalid parameter (%s), must be in x.x format", variable)
			}
		}

		storedValue, ok := variables[variable]
		if !ok {
			if len(dotKeys) == 0 {
				variables[variable] = value
			} else {
				// evaluate dot params and build map
				// last element in dot params will have value "value", the rest of map nested
				// using remaining elements
				// Ex: A.B.C = Z --> map[A:map[B:map[C:Z]]]
				variables[variable] = map[string]interface{}{dotKeys[len(dotKeys)-1]: value}
				for i := len(dotKeys) - 2; i > 0; i-- {
					variables[variable] = map[string]interface{}{dotKeys[i]: variables[variable]}
				}
			}
		} else {
			// if variable:value -> variable:[old value, newvalue]
			// if variable:[value1, value2] -> variable:[value1, value2, newvalue]
			// if map[variable[map[subvar:value]]] -> can't do anything, error
			switch storedValType := storedValue.(type) {
			case string:
				variables[variable] = []string{storedValType, value}
			case []string:
				variables[variable] = append(storedValType, value)
			case map[string]interface{}:
				return nil, fmt.Errorf(`(%s) was already used as an object path with the dot syntax
				 							and cannot be mixed with other formats`, variable)
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
