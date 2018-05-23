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

package svcattest

import (
	"io"
	"io/ioutil"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/spf13/viper"
)

// NewContext creates a test context for the svcat cli, optionally capturing the
// command output, or injecting a fake set of clients.
func NewContext(outputCapture io.Writer, fakeApp *svcat.App) *command.Context {
	if outputCapture == nil {
		outputCapture = ioutil.Discard
	}

	return &command.Context{
		Viper:  viper.New(),
		Output: outputCapture,
		App:    fakeApp,
	}
}
