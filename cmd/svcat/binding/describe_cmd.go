package binding

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	*command.Context
	ns       string
	name     string
	traverse bool
}

// NewDescribeCmd builds a "svcat describe binding" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "binding NAME",
		Aliases: []string{"bindings", "bnd"},
		Short:   "Show details of a specific binding",
		Example: `
  svcat describe binding wordpress-mysql-binding
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return describeCmd.run(args)
		},
	}
	cmd.Flags().StringVarP(
		&describeCmd.ns,
		"namespace",
		"n",
		"default",
		"The namespace in which to get the binding",
	)
	cmd.Flags().BoolVarP(
		&describeCmd.traverse,
		"traverse",
		"t",
		false,
		"Whether or not to traverse from binding -> instance -> class/plan -> broker",
	)
	return cmd
}

func (c *describeCmd) run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("name is required")
	}
	c.name = args[0]

	return c.describe()
}

func (c *describeCmd) describe() error {
	binding, err := c.App.RetrieveBinding(c.ns, c.name)
	if err != nil {
		return err
	}

	output.WriteBindingDetails(c.Output, binding)

	if c.traverse {
		instance, class, plan, broker, err := c.App.BindingParentHierarchy(binding)
		if err != nil {
			return fmt.Errorf("unable to traverse up the binding hierarchy (%s)", err)
		}
		output.WriteParentInstance(c.Output, instance)
		output.WriteParentClass(c.Output, class)
		output.WriteParentPlan(c.Output, plan)
		output.WriteParentBroker(c.Output, broker)
	}

	return nil
}
