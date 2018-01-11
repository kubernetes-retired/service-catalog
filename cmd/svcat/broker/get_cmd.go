package broker

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

type getCmd struct {
	*command.Context
	name string
}

// NewGetCmd builds a "svcat get brokers" command
func NewGetCmd(cxt *command.Context) *cobra.Command {
	getCmd := getCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "brokers [name]",
		Aliases: []string{"broker", "brk"},
		Short:   "List brokers, optionally filtered by name",
		Example: `
  svcat get brokers
  svcat get broker asb
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getCmd.run(args)
		},
	}

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
	brokers, err := c.App.RetrieveBrokers()
	if err != nil {
		return err
	}

	output.WriteBrokerList(c.Output, brokers...)
	return nil
}

func (c *getCmd) get() error {
	broker, err := c.App.RetrieveBroker(c.name)
	if err != nil {
		return err
	}

	output.WriteBrokerList(c.Output, *broker)
	return nil
}
