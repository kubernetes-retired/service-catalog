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
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

type deprovisonCmd struct {
	*command.Namespaced
	instanceName string
	wait         bool
	rawTimeout   string
	timeout      *time.Duration
}

// NewDeprovisionCmd builds a "svcat deprovision" command
func NewDeprovisionCmd(cxt *command.Context) *cobra.Command {
	deprovisonCmd := &deprovisonCmd{Namespaced: command.NewNamespacedCommand(cxt)}
	cmd := &cobra.Command{
		Use:   "deprovision NAME",
		Short: "Deletes an instance of a service",
		Example: `
  svcat deprovision wordpress-mysql-instance
`,
		PreRunE: command.PreRunE(deprovisonCmd),
		RunE:    command.RunE(deprovisonCmd),
	}
	command.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().BoolVar(&deprovisonCmd.wait, "wait", false,
		"Wait until the operation completes.")
	cmd.Flags().StringVar(&deprovisonCmd.rawTimeout, "timeout", "5m",
		"Timeout for --wait, specified in human readable format: 30s, 1m, 1h. Specify -1 to wait indefinitely.")

	return cmd
}

func (c *deprovisonCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("name is required")
	}
	c.instanceName = args[0]

	if c.wait && c.rawTimeout != "-1" {
		timeout, err := time.ParseDuration(c.rawTimeout)
		if err != nil {
			return fmt.Errorf("invalid --timeout value (%s)", err)
		}
		c.timeout = &timeout
	}

	return nil
}

func (c *deprovisonCmd) Run() error {
	return c.deprovision()
}

func (c *deprovisonCmd) deprovision() error {
	err := c.App.Deprovision(c.Namespace, c.instanceName)
	if err != nil {
		return err
	}

	if c.wait {
		glog.V(2).Infof("Waiting for the instance to be deprovisioned...")
		pollInterval := 1 * time.Second

		var instance *v1beta1.ServiceInstance
		instance, err = c.App.WaitForInstance(c.Namespace, c.instanceName, pollInterval, c.timeout)

		// The instance failed to deprovision cleanly, dump out more information on why
		if c.App.IsInstanceFailed(instance) {
			output.WriteInstanceDetails(c.Output, instance)
		}
	}

	if err == nil {
		output.WriteDeletedResourceName(c.Output, c.instanceName)
	}
	return err
}
