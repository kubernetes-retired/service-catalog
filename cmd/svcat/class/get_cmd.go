package class

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

type getCmd struct {
	*command.Context
	lookupByUUID bool
	uuid         string
	name         string
}

// NewGetCmd builds a "svcat get classes" command
func NewGetCmd(cxt *command.Context) *cobra.Command {
	getCmd := &getCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "classes [name]",
		Aliases: []string{"class", "cl"},
		Short:   "List classes, optionally filtered by name",
		Example: `
  svcat get classes
  svcat get class azure-mysqldb
  svcat get class --uuid 997b8372-8dac-40ac-ae65-758b4a5075a5
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getCmd.run(args)
		},
	}
	cmd.Flags().BoolVarP(
		&getCmd.lookupByUUID,
		"uuid",
		"u",
		false,
		"Whether or not to get the class by UUID (the default is by name)",
	)
	return cmd
}

func (c *getCmd) run(args []string) error {
	if len(args) == 0 {
		return c.getAll()
	}

	if c.lookupByUUID {
		c.uuid = args[0]
	} else {
		c.name = args[0]
	}

	return c.get()
}

func (c *getCmd) getAll() error {
	classes, err := c.App.RetrieveClasses()
	if err != nil {
		return err
	}

	output.WriteClassList(c.Output, classes...)
	return nil
}

func (c *getCmd) get() error {
	var class *v1beta1.ClusterServiceClass
	var err error

	if c.lookupByUUID {
		class, err = c.App.RetrieveClassByID(c.uuid)
	} else if c.name != "" {
		class, err = c.App.RetrieveClassByName(c.name)
	}
	if err != nil {
		return err
	}

	output.WriteClassList(c.Output, *class)
	return nil
}
