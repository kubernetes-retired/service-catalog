package broker

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

type syncCmd struct {
	*command.Context
	name string
}

// NewSyncCmd builds a "svcat sync broker" command
func NewSyncCmd(cxt *command.Context) *cobra.Command {
	syncCmd := syncCmd{Context: cxt}
	rootCmd := &cobra.Command{
		Use:   "broker [name]",
		Short: "Syncs service catalog for a service broker",
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncCmd.run(args)
		},
	}
	return rootCmd
}

func (c *syncCmd) run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("name is required")
	}
	c.name = args[0]
	return c.sync()
}

func (c *syncCmd) sync() error {
	const retries = 3
	err := c.App.Sync(c.name, retries)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.Output, "Successfully fetched catalog entries from the %s broker\n", c.name)
	return nil
}
