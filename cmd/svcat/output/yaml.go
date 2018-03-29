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

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime"
)

// writeYAML writes the given obj to the given Writer in YAML format, indented
// n spaces
func writeYAML(w io.Writer, obj interface{}, n int) {
	yBytes, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Fprintf(w, "err marshaling yaml: %v\n", err)
		return
	}
	y := string(yBytes)
	if n > 0 {
		indent := strings.Repeat(" ", n)
		y = indent + strings.Replace(y, "\n", "\n"+indent, -1)
		y = strings.TrimRight(y, " ")
	}

	fmt.Fprint(w, y)
}

func writeParameters(w io.Writer, parameters *runtime.RawExtension) {
	if parameters == nil {
		return
	}

	fmt.Fprintln(w, "\nParameters:")
	var params map[string]interface{}
	err := json.Unmarshal(parameters.Raw, &params)
	if err != nil {
		// If it isn't formatted in json, just show the string representation of what is present
		fmt.Fprintln(w, string(parameters.Raw))
	} else {
		writeYAML(w, params, 2)
	}
}
