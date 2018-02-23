package touch

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

// NewTouchCmd builds a "svcat touch [instance, binding]"
func NewTouchCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "touch instance",
		Short:   "Make the service catalog attempt to re-provision an instance",
		Example: "svcat touch instance wordpress-mysql-instance",
	}
	cmd.AddCommand(newTouchInstanceCmd(cxt))
	return cmd
}
