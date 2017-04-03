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

package wholepkg

// Trivial
type Struct_Empty struct{}

// Only primitives
type Struct_Primitives struct {
	BoolField   bool
	IntField    int
	StringField string
	FloatField  float64
}
type Struct_Primitives_Alias Struct_Primitives
type Struct_Embed_Struct_Primitives struct {
	Struct_Primitives
}
type Struct_Embed_Int struct {
	int
}
type Struct_Struct_Primitives struct {
	StructField Struct_Primitives
}

// Manual DeepCopy method
type ManualStruct struct {
	StringField string
}

func (m ManualStruct) DeepCopy() ManualStruct {
	return m
}

// Everything
type Struct_Everything struct {
	BoolField         bool
	IntField          int
	StringField       string
	FloatField        float64
	StructField       Struct_Primitives
	EmptyStructField  Struct_Empty
	ManualStructField ManualStruct
}

/*
// Only pointers to primitives
type Struct_PrimitivePointers struct {
	BoolField   *bool
	IntField    *int
	StringField *string
	FloatField  *float64
}
*/
