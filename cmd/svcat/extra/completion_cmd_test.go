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

package extra_test

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	. "github.com/kubernetes-incubator/service-catalog/cmd/svcat/extra"
	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Completion Command", func() {
	Describe("NewCompletionCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewCompletionCmd(cxt)
			Expect(*cmd).NotTo(BeNil())

			Expect(cmd.Use).To(Equal("completion SHELL"))
			Expect(cmd.Short).To(ContainSubstring("Output shell completion code for the specified shell"))
		})
	})
	Describe("Validate", func() {
		It("succeeds if a shell name is provided", func() {
			cmd := CompletionCmd{}
			err := cmd.Validate([]string{"bash"})
			Expect(err).NotTo(HaveOccurred())

			cmd = CompletionCmd{}
			err = cmd.Validate([]string{"zsh"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("errors if a shell name is not provided", func() {
			cmd := CompletionCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Shell not specified"))
		})
		It("errors if provided the name of an unsupported shell", func() {
			cmd := CompletionCmd{}
			err := cmd.Validate([]string{"csh"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unsupported shell type"))
		})
	})
	Describe("Run", func() {
		It("floops the pig", func() {
		})
	})
})
