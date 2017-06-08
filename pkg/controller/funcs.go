/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"runtime/debug"
	"testing"
)

// FailfFunc is a type that defines the common signatures of T.Fatalf and T.Errorf
type FailfFunc func(t *testing.T, msg string, args ...interface{})

// Fatalf is a FailfFunc that logs a stack trace and then calls t.Fatalf
func Fatalf(t *testing.T, msg string, args ...interface{}) {
	t.Log(string(debug.Stack()))
	t.Fatalf(msg, args...)
}

// Errorf is a FailfFunc that logs a stack trace and then calls t.Errorf
func Errorf(t *testing.T, msg string, args ...interface{}) {
	t.Log(string(debug.Stack()))
	t.Errorf(msg, args...)
}
