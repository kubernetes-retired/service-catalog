package instance

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

// NewDescribeCmd builds a "svcat describe instance" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "instance NAME",
		Aliases: []string{"instances", "inst"},
		Short:   "Show details of a specific instance",
		Example: `
  svcat describe instance wordpress-mysql-instance
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
		"The namespace in which to get the instance",
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
	instance, err := c.App.RetrieveInstance(c.ns, c.name)
	if err != nil {
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)

	bindings, err := c.App.RetrieveBindingsByInstance(instance)
	if err != nil {
		return err
	}
	output.WriteAssociatedBindings(c.Output, bindings)

	if c.traverse {
		class, plan, broker, err := c.App.InstanceParentHierarchy(instance)
		if err != nil {
			return fmt.Errorf("unable to traverse up the instance hierarchy (%s)", err)
		}
		output.WriteParentClass(c.Output, class)
		output.WriteParentPlan(c.Output, plan)
		output.WriteParentBroker(c.Output, broker)
	}

	return nil
}
