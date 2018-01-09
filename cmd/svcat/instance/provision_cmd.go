package instance

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/parameters"
	"github.com/spf13/cobra"
)

type provisonCmd struct {
	*command.Context
	ns           string
	instanceName string
	className    string
	planName     string
	rawParams    []string
	params       map[string]string
	rawSecrets   []string
	secrets      map[string]string
}

// NewProvisionCmd builds a "svcat provision" command
func NewProvisionCmd(cxt *command.Context) *cobra.Command {
	provisionCmd := &provisonCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "provision NAME --plan PLAN --class CLASS",
		Short: "Create a new instance of a service",
		Example: `
  svcat provision wordpress-mysql-instance --class azure-mysqldb --plan standard800 -p location=eastus -p sslEnforcement=disabled
  svcat provision wordpress-mysql-instance --class azure-mysqldb --plan standard800 -s mysecret[dbparams]
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return provisionCmd.run(args)
		},
	}
	cmd.Flags().StringVarP(&provisionCmd.ns, "namespace", "n", "default",
		"The namespace in which to create the instance")
	cmd.Flags().StringVar(&provisionCmd.className, "class", "",
		"The class name (Required)")
	cmd.MarkFlagRequired("class")
	cmd.Flags().StringVar(&provisionCmd.planName, "plan", "",
		"The plan name (Required)")
	cmd.MarkFlagRequired("plan")
	cmd.Flags().StringArrayVarP(&provisionCmd.rawParams, "param", "p", nil,
		"Additional parameter to use when provisioning the service, format: NAME=VALUE")
	cmd.Flags().StringArrayVarP(&provisionCmd.rawSecrets, "secret", "s", nil,
		"Additional parameter, whose value is stored in a secret, to use when provisioning the service, format: SECRET[KEY]")
	return cmd
}

func (c *provisonCmd) run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.instanceName = args[0]

	var err error

	c.params, err = parameters.ParseVariableAssignments(c.rawParams)
	if err != nil {
		return fmt.Errorf("invalid --param value (%s)", err)
	}

	c.secrets, err = parameters.ParseKeyMaps(c.rawSecrets)
	if err != nil {
		return fmt.Errorf("invalid --secret value (%s)", err)
	}

	return c.provision()
}

func (c *provisonCmd) provision() error {
	instance, err := c.App.Provision(c.ns, c.instanceName, c.className, c.planName, c.params, c.secrets)
	if err != nil {
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)

	return nil
}
