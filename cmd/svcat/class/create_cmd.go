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
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// CreateCmd contains the information needed to create a new class
type CreateCmd struct {
	*command.Context
	Name      string
	From      string
	Namespace string
	Cluster   bool
}

// NewCreateCmd builds a "svcat create class" command
func NewCreateCmd(cxt *command.Context) *cobra.Command {
	createCmd := &CreateCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "class [NAME] --from [EXISTING_NAME] [--namespace [NAMESPACE]|--cluster]",
		Short: "Copies an existing class into a new user-defined cluster-scoped class",
		Example: command.NormalizeExamples(`
  svcat create class newclass --from mysqldb
  svcat create class newclass --from mysqldb --cluster
  svcat create class newclass --from mysqldb --namespace newnamespace
`),
		PreRunE: command.PreRunE(createCmd),
		RunE:    command.RunE(createCmd),
	}
	cmd.Flags().StringVarP(&createCmd.From, "from", "f", "",
		"Name from an existing class that will be copied (Required)",
	)
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringVarP(&createCmd.Namespace, "namespace", "n", "",
		"Overrides the current namespace to specify the namespace of the namespace-scoped class",
	)
	cmd.Flags().BoolVarP(&createCmd.Cluster, "cluster", "c", false,
		"Specifies that a cluster-scoped class should be created. By default namespace-scoped classes are created. When this flag is set, namespace is ignored.",
	)

	return cmd
}

// Validate checks that the required arguments have been provided
func (c *CreateCmd) Validate(args []string) error {
	if len(args) <= 0 {
		return fmt.Errorf("new class name should be provided")
	}

	if c.Namespace != "" && c.Cluster {
		return fmt.Errorf("namespace can not be specified for a cluster-scoped class")
	}

	c.Name = args[0]

	return nil
}

// Run calls out to the pkg lib to create the class and displays the output
func (c *CreateCmd) Run() error {
	opts := servicecatalog.CreateClassFromOptions{
		Name:      c.Name,
		Namespace: c.Namespace,
		From:      c.From,
		Scope:     servicecatalog.NamespaceScope,
	}

	if c.Cluster {
		opts.Scope = servicecatalog.ClusterScope
	}

	createdClass, err := c.App.CreateClassFrom(opts)
	if err != nil {
		return err
	}

	output.WriteClassList(c.Output, output.FormatTable, createdClass)
	return nil
}
