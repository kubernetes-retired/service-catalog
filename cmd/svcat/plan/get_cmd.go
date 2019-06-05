/*
Copyright 2019 The Kubernetes Authors.

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
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// GetCmd contains the information needed to get a specific plan or all plans
type GetCmd struct {
	*command.Namespaced
	*command.Scoped
	*command.Formatted
	LookupByKubeName bool
	KubeName         string
	Name             string

	ClassFilter   string
	ClassKubeName string
	ClassName     string
}

// NewGetCmd builds a "svcat get plans" command
func NewGetCmd(ctx *command.Context) *cobra.Command {
	getCmd := &GetCmd{
		Namespaced: command.NewNamespaced(ctx),
		Scoped:     command.NewScoped(),
		Formatted:  command.NewFormatted(),
	}
	cmd := &cobra.Command{
		Use:     "plans [NAME]",
		Aliases: []string{"plan", "pl"},
		Short:   "List plans, optionally filtered by name, class, scope or namespace",
		Example: command.NormalizeExamples(`
  svcat get plans
  svcat get plans --scope cluster
  svcat get plans --scope namespace --namespace dev
  svcat get plan PLAN_NAME
  svcat get plan CLASS_NAME/PLAN_NAME
  svcat get plan --kube-name PLAN_KUBE_NAME
  svcat get plans --class CLASS_NAME
  svcat get plan --class CLASS_NAME PLAN_NAME
  svcat get plans --kube-name --class CLASS_KUBE_NAME
  svcat get plan --kube-name --class CLASS_KUBE_NAME PLAN_KUBE_NAME
`),
		PreRunE: command.PreRunE(getCmd),
		RunE:    command.RunE(getCmd),
	}
	cmd.Flags().BoolVarP(
		&getCmd.LookupByKubeName,
		"kube-name",
		"k",
		false,
		"Whether or not to get the plan by its Kubernetes name (the default is by external name)",
	)
	cmd.Flags().StringVarP(
		&getCmd.ClassFilter,
		"class",
		"c",
		"",
		"Filter plans based on class. When --kube-name is specified, the class name is interpreted as a kubernetes name.",
	)
	getCmd.AddOutputFlags(cmd.Flags())
	getCmd.AddNamespaceFlags(cmd.Flags(), true)
	getCmd.AddScopedFlags(cmd.Flags(), true)
	return cmd
}

// Validate parses the provided arugments and errors if they are formatted incorrectly
func (c *GetCmd) Validate(args []string) error {
	if len(args) > 0 {
		if c.LookupByKubeName {
			if strings.Contains(args[0], "/") {
				names := strings.Split(args[0], "/")
				if len(names) != 2 {
					return fmt.Errorf("failed to parse class/plan k8s name combination '%s'", args[0])
				}
				c.ClassKubeName = names[0]
				c.KubeName = names[1]
			} else {
				c.KubeName = args[0]
			}
		} else if strings.Contains(args[0], "/") {
			names := strings.Split(args[0], "/")
			if len(names) != 2 {
				return fmt.Errorf("failed to parse class/plan name combination '%s'", args[0])
			}
			c.ClassName = names[0]
			c.Name = names[1]
		} else {
			c.Name = args[0]
		}
	}
	if c.ClassFilter != "" {
		if c.LookupByKubeName {
			c.ClassKubeName = c.ClassFilter
		} else {
			c.ClassName = c.ClassFilter
		}
	}

	return nil
}

// Run determines if we are fetching all plans or a specific one, and calls
// the corresponding method
func (c *GetCmd) Run() error {
	if c.KubeName == "" && c.Name == "" {
		return c.getAll()
	}

	return c.get()
}

func (c *GetCmd) getAll() error {
	// Retrieve the classes as well because plans don't have the external class name
	classOpts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	classes, err := c.App.RetrieveClasses(classOpts)
	if err != nil {
		return fmt.Errorf("unable to list classes (%s)", err)
	}

	var classID string
	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	if c.ClassFilter != "" {
		if !c.LookupByKubeName {
			// Map the external class name to the class name.
			for _, class := range classes {
				if c.ClassName == class.GetExternalName() {
					c.ClassKubeName = class.GetName()
					break
				}
			}
		}
		classID = c.ClassKubeName
	}

	plans, err := c.App.RetrievePlans(classID, opts)
	if err != nil {
		return fmt.Errorf("unable to list plans (%s)", err)
	}
	output.WritePlanList(c.Output, c.OutputFormat, plans, classes)
	return nil
}

func (c *GetCmd) get() error {
	var plan servicecatalog.Plan
	var err error

	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}

	switch {
	case c.LookupByKubeName:
		plan, err = c.App.RetrievePlanByID(c.KubeName, opts)

	case c.ClassName != "":
		plan, err = c.App.RetrievePlanByClassAndName(c.ClassName, c.Name, opts)

	default:
		plan, err = c.App.RetrievePlanByName(c.Name, opts)

	}
	if err != nil {
		return err
	}
	// Retrieve the class as well because plans don't have the external class name
	class, err := c.App.RetrieveClassByID(plan.GetClassID(), opts)
	if err != nil {
		return err
	}

	output.WritePlan(c.Output, c.OutputFormat, plan, class)

	return nil
}
