/*
Copyright 2016 The Kubernetes Authors.

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

package mem

import (
	"bytes"
	"encoding/gob"
)

// deepCopy is a quick and dirty method for creating a deep of any arbitrary
// type. It does this by serializing the original to a byte array, then
// deserializing those bytes. This is not necessarily very efficient, but the
// contents of this package are used only to enable tests, so the small bit
// of performance overhead has been judged tolerable.
func deepCopy(copy, orig interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(orig)
	if err != nil {
		return err
	}
	return dec.Decode(copy)
}
