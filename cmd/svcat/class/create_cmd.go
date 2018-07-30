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

package class

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

// CreateCmd contains the information needed to create a new class
type CreateCmd struct {
	*command.Context
	Name string
	From string
}

// NewCreateCmd builds a "svcat create class" command
func NewCreateCmd(cxt *command.Context) *cobra.Command {
	createCmd := &CreateCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "class [NAME] --from [EXISTING_NAME]",
		Short: "Copies an existing class into a new user-defined cluster-scoped class",
		Example: command.NormalizeExamples(`
  svcat create class newclass --from mysqldb
`),
		PreRunE: command.PreRunE(createCmd),
		RunE:    command.RunE(createCmd),
	}
	cmd.Flags().StringVarP(&createCmd.From, "from", "f", "",
		"Name from an existing class that will be copied (Required)",
	)
	cmd.MarkFlagRequired("from")

	return cmd
}

// Validate checks that the required arguments have been provided
func (c *CreateCmd) Validate(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("new class name should be provided")
	}

	c.Name = args[0]

	return nil
}

// Run calls out to the pkg lib to create the class and displays the output
func (c *CreateCmd) Run() error {
	class, err := c.App.RetrieveClassByName(c.From)
	if err != nil {
		return err
	}

	class.Name = c.Name

	createdClass, err := c.App.CreateClass(class)
	if err != nil {
		return err
	}

	output.WriteClassDetails(c.Output, createdClass)
	return nil
}
