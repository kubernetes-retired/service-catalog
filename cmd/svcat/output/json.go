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
)

func writeJSON(w io.Writer, obj interface{}) {
	indent := strings.Repeat(" ", 3)
	j, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		fmt.Fprintf(w, "err marshaling json: %v\n", err)
		return
	}
	fmt.Fprint(w, string(j))
}
