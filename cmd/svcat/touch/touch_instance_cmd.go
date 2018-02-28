package touch

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

type touchInstanceCmd struct {
	*command.Context
	namespace string
	name      string
}

func newTouchInstanceCmd(cxt *command.Context) *cobra.Command {
	touchInstanceCmd := &touchInstanceCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "touch an instance to make service-catalog try to provision again",
		Example: `
svcat touch instance wordpress-mysql-instance --namespace mynamespace
`,
		PreRunE: command.PreRunE(touchInstanceCmd),
		RunE:    command.RunE(touchInstanceCmd),
	}
	cmd.Flags().StringVarP(&touchInstanceCmd.namespace, "namespace", "n", "default",
		"The namespace for the instance to touch")
	return cmd
}

func (c *touchInstanceCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.name = args[0]

	return nil
}

func (c *touchInstanceCmd) Run() error {
	const retries = 3
	return c.App.TouchInstance(c.namespace, c.name, retries)
}
