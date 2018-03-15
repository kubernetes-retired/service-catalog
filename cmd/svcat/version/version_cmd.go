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

package binding

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/parameters"
	"github.com/spf13/cobra"
)

type versionCmd struct {
	*command.Context
}

// NewVersionCmd builds a "svcat version" command
func NewVersionCmd(cxt *command.Context) *cobra.Command {
	versionCmd := &versionCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Provides the version for the Service Catalog client and server",
		Example: `
  svcat version
  svcat version --client
  svcat version --server
`,
		PreRunE: command.PreRunE(versionCmd),
		RunE:    command.RunE(versionCmd),
	}
	cmd.Flags().StringVarP(
		&versionCmd.ns,
		"namespace",
		"n",
		"default",
		"The instance namespace",
	)
	cmd.Flags().StringVarP(
		&bindCmd.bindingName,
		"name",
		"",
		"",
		"The name of the binding. Defaults to the name of the instance.",
	)
	cmd.Flags().StringVarP(
		&bindCmd.secretName,
		"secret-name",
		"",
		"",
		"The name of the secret. Defaults to the name of the instance.",
	)
	cmd.Flags().StringSliceVarP(&bindCmd.rawParams, "param", "p", nil,
		"Additional parameter to use when binding the instance, format: NAME=VALUE")
	cmd.Flags().StringSliceVarP(&bindCmd.rawSecrets, "secret", "s", nil,
		"Additional parameter, whose value is stored in a secret, to use when binding the instance, format: SECRET[KEY]")

	return cmd
}

func (c *versionCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("instance is required")
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

	return nil
}

func (c *versionCmd) Run() error {
	return c.bind()
}

func (c *versionCmd) bind() error {
	binding, err := c.App.Bind(c.ns, c.bindingName, c.instanceName, c.secretName, c.params, c.secrets)
	if err != nil {
		return err
	}

	output.WriteBindingDetails(c.Output, binding)
	return nil
}
