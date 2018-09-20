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

package plan

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

// CreateCmd contains the information needed to create a new plan.
type CreateCmd struct {
	*command.Context
	Name string
	From string
}

// NewCreatedCmd builds a "svcat create plan" command.
func NewCreateCmd(ctx *command.Context) *cobra.Command {
	createCmd := &CreateCmd{
		Context: ctx,
	}

	cmd := &cobra.Command{
		Use:   "plan [NAME] --from [EXISTING_NAME]",
		Short: "Copies an existing plan into a new user-defined cluster-scoped plan",
		Example: command.NormalizeExamples(`
svcat create plan newplan --from mysqldb
`),
		PreRunE: command.PreRunE(createCmd),
		RunE:    command.RunE(createCmd),
	}
	cmd.Flags().StringVarP(&createCmd.From, "from", "f", "",
		"Name of an existing class that will be copied (Required)",
	)
	cmd.MarkFlagRequired("from")
	return cmd
}

// Validate checks that the required arguments have been passed.
func (c *CreateCmd) Validate(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("new class name should be provided")
	}
	c.Name = args[0]
	return nil
}

// Run calls the pkg lib to create a plan and displays the output.
func (c *CreateCmd) Run() error {
	var err error
	plan, err := c.App.RetrievePlanByName(c.From)
	if err != nil {
		return err
	}
	plan.Name = c.Name
	createdPlan, err := c.App.CreatePlan(plan)
	if err != nil {
		return err
	}
	class, err := c.App.RetrieveClassByPlan(createdPlan)
	if err != nil {
		return err
	}
	output.WritePlanDetails(c.Output, createdPlan, class)
	return nil
}
