package binding

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

type getCmd struct {
	*command.Context
	ns   string
	name string
}

// NewGetCmd builds a "svcat get bindings" command
func NewGetCmd(cxt *command.Context) *cobra.Command {
	getCmd := getCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "bindings [name]",
		Aliases: []string{"binding", "bnd"},
		Short:   "List bindings, optionally filtered by name",
		Example: `
  svcat get bindings
  svcat get binding wordpress-mysql-binding
  svcat get binding -n ci concourse-postgres-binding
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getCmd.run(args)
		},
	}

	cmd.Flags().StringVarP(
		&getCmd.ns,
		"namespace",
		"n",
		"default",
		"The namespace from which to get the bindings",
	)
	return cmd
}

func (c *getCmd) run(args []string) error {
	if len(args) == 0 {
		return c.getAll()
	}

	c.name = args[0]
	return c.get()
}

func (c *getCmd) getAll() error {
	bindings, err := c.App.RetrieveBindings(c.ns)
	if err != nil {
		return err
	}

	output.WriteBindingList(c.Output, bindings.Items...)
	return nil
}

func (c *getCmd) get() error {
	binding, err := c.App.RetrieveBinding(c.ns, c.name)
	if err != nil {
		return err
	}

	output.WriteBindingList(c.Output, *binding)
	return nil
}
