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

package command

import (
	"github.com/spf13/cobra"
)

// HasClassFlag represents a command that supports --class.
type HasClassFlag interface {
	// ApplyClassFlag validates and persists the class related flag.
	//   --class
	ApplyClassFlag(*cobra.Command) error
}

// ClassFilteredCommand adds support to a command for the --class flag.
type ClassFilteredCommand struct {
	ClassFilter string
}

// NewClassFilteredCommand initializes a new class specified command.
func NewClassFilteredCommand() *ClassFilteredCommand {
	return &ClassFilteredCommand{}
}

// AddClassFlag adds the class related flag.
//   --class
func (c *ClassFilteredCommand) AddClassFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"class",
		"c",
		"",
		"If present, specify the class used as a filter for this request",
	)
}

// ApplyClassFlag persists the class related flag.
//   --class
func (c *ClassFilteredCommand) ApplyClassFlag(cmd *cobra.Command) error {
	var err error
	c.ClassFilter, err = cmd.Flags().GetString("class")
	return err
}
