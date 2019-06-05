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
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// GetCmd contains the information needed to Get a specific class or all classes
type GetCmd struct {
	*command.Namespaced
	*command.Scoped
	*command.Formatted

	LookupByKubeName bool
	KubeName         string
	Name             string
}

// NewGetCmd builds a "svcat get classes" command
func NewGetCmd(cxt *command.Context) *cobra.Command {
	getCmd := &GetCmd{
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
		Formatted:  command.NewFormatted(),
	}
	cmd := &cobra.Command{
		Use:     "classes [NAME]",
		Aliases: []string{"class", "cl"},
		Short:   "List classes, optionally filtered by name, scope or namespace",
		Example: command.NormalizeExamples(`
  svcat get classes
  svcat get classes --scope cluster
  svcat get classes --scope namespace --namespace dev
  svcat get class mysqldb
  svcat get class --kube-name 997b8372-8dac-40ac-ae65-758b4a5075a5
`),
		PreRunE: command.PreRunE(getCmd),
		RunE:    command.RunE(getCmd),
	}
	cmd.Flags().BoolVarP(
		&getCmd.LookupByKubeName,
		"kube-name",
		"k",
		false,
		"Whether or not to get the class by its Kubernetes name (the default is by external name)",
	)
	getCmd.AddOutputFlags(cmd.Flags())
	getCmd.AddNamespaceFlags(cmd.Flags(), true)
	getCmd.AddScopedFlags(cmd.Flags(), true)
	return cmd
}

// Validate checks that the required arguments have been provided
func (c *GetCmd) Validate(args []string) error {
	if len(args) > 0 {
		if c.LookupByKubeName {
			c.KubeName = args[0]
		} else {
			c.Name = args[0]
		}
	}

	return nil
}

// Run determines if we're getting a single class or all classes,
// and calls the pertinent function
func (c *GetCmd) Run() error {
	if c.KubeName == "" && c.Name == "" {
		return c.getAll()
	}

	return c.get()
}

func (c *GetCmd) getAll() error {
	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	classes, err := c.App.RetrieveClasses(opts)
	if err != nil {
		return err
	}
	output.WriteClassList(c.Output, c.OutputFormat, classes...)
	return nil
}

func (c *GetCmd) get() error {
	var class servicecatalog.Class
	var err error
	scopeOpts := servicecatalog.ScopeOptions{
		Scope:     c.Scope,
		Namespace: c.Namespace,
	}
	if c.LookupByKubeName {
		class, err = c.App.RetrieveClassByID(c.KubeName, scopeOpts)
	} else if c.Name != "" {
		class, err = c.App.RetrieveClassByName(c.Name, scopeOpts)
	}
	if err != nil {
		if strings.Contains(err.Error(), servicecatalog.MultipleClassesFoundError) {
			return fmt.Errorf(err.Error() + ", please specify a scope with --scope or an exact Kubernetes name with --kube-name")
		}
		return err
	}

	output.WriteClass(c.Output, c.OutputFormat, class)
	return nil
}
