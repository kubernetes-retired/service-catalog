/*
Copyright 2018 The Kubernetes Authors.

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

package instance

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/parameters"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

type setCmd struct {
	*command.Namespaced
	*command.Waitable

	instanceName      string
	planName          string
	rawParams         []string
	jsonParams        string
	params            interface{}
	areParamsProvided bool
	rawSecrets        []string
	secrets           map[string]string
}

// NewSetCmd builds a "svcat set instance" command
func NewSetCmd(cxt *command.Context) *cobra.Command {
	setCmd := &setCmd{
		Namespaced: command.NewNamespaced(cxt),
		Waitable:   command.NewWaitable(),
	}
	cmd := &cobra.Command{
		Use:   "instance NAME [flags]",
		Short: "Configure a provisioned service instance",
		Example: command.NormalizeExamples(`
  svcat set instance wordpress-mysql-instance --plan free
  svcat set instance wordpress-mysql-instance --plan free -s mysecret[dbparams]
  svcat set instance secure-instance --plan secureDB --params-json '{
    "encrypt" : true,
    "firewallRules" : [
        {
            "name": "AllowSome",
            "startIPAddress": "75.70.113.50",
            "endIPAddress" : "75.70.113.131"
        }
    ]
  }'
`),
		PreRunE: command.PreRunE(setCmd),
		RunE:    command.RunE(setCmd),
	}
	setCmd.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().StringVar(&setCmd.planName, "plan", "",
		"The plan name (Required)")
	cmd.Flags().StringSliceVarP(&setCmd.rawParams, "param", "p", nil,
		"Additional parameter to use when updating the service instance, format: NAME=VALUE. Cannot be combined with --params-json, Sensitive information should be placed in a secret and specified with --secret")
	cmd.Flags().StringSliceVarP(&setCmd.rawSecrets, "secret", "s", nil,
		"Additional parameter, whose value is stored in a secret, to use when updating the service instance, format: SECRET[KEY]")
	cmd.Flags().StringVar(&setCmd.jsonParams, "params-json", "",
		"Additional parameters to use when updating the service instance, provided as a JSON object. Cannot be combined with --param")
	setCmd.AddWaitFlags(cmd)

	return cmd
}

func (c *setCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.instanceName = args[0]

	var err error

	if c.jsonParams != "" && len(c.rawParams) > 0 {
		return fmt.Errorf("--params-json cannot be used with --param")
	}

	if c.jsonParams != "" {
		c.params, err = parameters.ParseVariableJSON(c.jsonParams)
		if err != nil {
			return fmt.Errorf("invalid --params-json value (%s)", err)
		}
		c.areParamsProvided = true
	} else if len(c.rawParams) > 0 {
		c.params, err = parameters.ParseVariableAssignments(c.rawParams)
		if err != nil {
			return fmt.Errorf("invalid --param value (%s)", err)
		}
		c.areParamsProvided = true
	}

	c.secrets, err = parameters.ParseKeyMaps(c.rawSecrets)
	if err != nil {
		return fmt.Errorf("invalid --secret value (%s)", err)
	}

	return nil
}

func (c *setCmd) Run() error {
	err := c.configureCmdProperties()
	if err != nil {
		return err
	}

	return c.Update()
}

func (c *setCmd) Update() error {
	instance, err := c.App.UpdateInstance(c.Namespace, c.instanceName, c.planName, c.params, c.secrets)
	if err != nil {
		return err
	}

	if c.Wait {
		fmt.Fprintln(c.Output, "Waiting for the instance to be updated...")
		finalInstance, err := c.App.WaitForInstance(instance.Namespace, instance.Name, c.Interval, c.Timeout)
		if err == nil {
			instance = finalInstance
		}

		// Always print the instance because the update did succeed,
		// and just print any errors that occurred while polling
		output.WriteInstanceDetails(c.Output, instance)
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)
	return nil
}

func (c *setCmd) configureCmdProperties() error {
	instance, err := c.App.RetrieveInstance(c.Namespace, c.instanceName)
	if err != nil {
		return nil
	}

	if c.planName == "" {
		c.planName = instance.Spec.PlanReference.ClusterServicePlanExternalName
	}

	if !c.areParamsProvided {
		c.params = servicecatalog.BuildParametersFromInstance(instance.Spec.Parameters)
	}

	if len(c.secrets) == 0 {
		c.secrets = servicecatalog.BuildMapFromInstanceSecretRefs(instance.Spec.ParametersFrom)
	}

	return nil
}
