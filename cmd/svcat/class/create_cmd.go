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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
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
	cmd.Flags().StringVarP(&createCmd.from, "from", "f", "",
		"Name from an existing class that will be copied (Required)",
	)
	cmd.MarkFlagRequired("from")

	return cmd
}

func (c *createCmd) Validate(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("new class name should be provided")
	}

	c.name = args[0]

	return nil
}

func (c *createCmd) Run() error {
	class, err := c.App.RetrieveClassByName(c.from)
	if err != nil {
		return err
	}

	newClass := &v1beta1.ClusterServiceClass{
		Spec: v1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: class.Spec.ClusterServiceBrokerName,
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: c.name,
				Description:  class.Spec.Description,
				Tags:         class.Spec.Tags,
			},
		},
	}

	createdClass, err := c.App.CreateClass(newClass)
	if err != nil {
		return err
	}

	output.WriteCreatedResourceName(c.Output, createdClass.Spec.ExternalName)
	return nil
}
