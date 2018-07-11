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
	"github.com/spf13/cobra"
)

type createCmd struct {
	*command.Context
	name string
	from string
}

// NewCreateCmd builds a "svcat create class" command
func NewCreateCmd(cxt *command.Context) *cobra.Command {
	createCmd := &createCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "class [NAME] --from [EXISTING_NAME]",
		Short: "Copies an existing class into a new user-defined cluster-scoped class",
		Example: command.NormalizeExamples(`
  svcat create class newclass --from mysqldb
`),
		PreRunE: command.PreRunE(createCmd),
		RunE:    command.RunE(createCmd),
	}
	cmd.Flags().StringVarP(
		&createCmd.from,
		"from",
		"f",
		"",
		"Name from an existing class that will be copied",
	)
	return cmd
}

func (c *createCmd) Validate(args []string) error {
	if len(args) > 0 {
		c.name = args[0]
	}

	if c.name == "" {
		return fmt.Errorf("new class name should be provided")
	}

	if c.from == "" {
		return fmt.Errorf("an exisitng class name should be provided")
	}

	return nil
}

func (c *createCmd) Run() error {
	class, err := c.App.RetrieveClassByName(c.from)
	if err != nil {
		return err
	}

	class.Spec.ExternalName = c.name

	_, err = c.App.CreateClass(class)
	if err != nil {
		return err
	}

	return nil
}
