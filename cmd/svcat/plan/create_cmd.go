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

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"

	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

// CreateCmd contains the information needed to create a new plan.
type CreateCmd struct {
	*command.Namespaced
	*command.Scoped
	Name string
	From string
}

// NewCreateCmd builds a "svcat create plan" command.
func NewCreateCmd(ctx *command.Context) *cobra.Command {
	createCmd := &CreateCmd{
		Namespaced: command.NewNamespaced(ctx),
		Scoped:     command.NewScoped(),
	}

	cmd := &cobra.Command{
		Use:   "plan [NAME] --from [EXISTING_NAME]",
		Short: "Copies an existing plan into a new user-defined cluster-scoped or namespace-scoped plan",
		Example: command.NormalizeExamples(`
svcat create plan newplan --from mysqldb
svcat create plan newplan --from mysqldb --scope cluster
  svcat create plan newplan --from mysqldb --scope namespace --namespace newnamespace
`),
		PreRunE: command.PreRunE(createCmd),
		RunE:    command.RunE(createCmd),
	}
	cmd.Flags().StringVarP(&createCmd.From, "from", "f", "",
		"Name of an existing plan that will be copied (Required)",
	)
	cmd.MarkFlagRequired("from")
	createCmd.Namespaced.AddNamespaceFlags(cmd.Flags(), true)
	createCmd.Scoped.AddScopedFlags(cmd.Flags(), true)
	return cmd
}

// Validate checks that the required arguments have been passed.
func (c *CreateCmd) Validate(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("new plan name should be provided")
	}
	c.Name = args[0]
	return nil
}

// Run calls the pkg lib to create a plan and displays the output.
func (c *CreateCmd) Run() error {
	opts := servicecatalog.CreatePlanFromOptions{
		Name:      c.Name,
		From:      c.From,
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	createdPlan, err := c.App.CreatePlan(opts)
	if err != nil {
		return err
	}
	className := createdPlan.GetClassID()
	class, err := c.App.RetrieveClassByName(className, servicecatalog.ScopeOptions{
		Scope:     c.Scope,
		Namespace: c.Namespace,
	})
	output.WritePlanDetails(c.Output, createdPlan, class)
	return nil
}
