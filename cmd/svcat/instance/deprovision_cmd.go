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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

type deprovisonCmd struct {
	*command.Namespaced
	*command.Waitable

	instanceName string
	abandon      bool
	skipPrompt   bool
}

// NewDeprovisionCmd builds a "svcat deprovision" command
func NewDeprovisionCmd(cxt *command.Context) *cobra.Command {
	deprovisonCmd := &deprovisonCmd{
		Namespaced: command.NewNamespaced(cxt),
		Waitable:   command.NewWaitable(),
	}
	cmd := &cobra.Command{
		Use:   "deprovision NAME",
		Short: "Deletes an instance of a service",
		Example: command.NormalizeExamples(`
  svcat deprovision wordpress-mysql-instance
  svcat deprovision --abandon wordpress-mysql-instance
`),
		PreRunE: command.PreRunE(deprovisonCmd),
		RunE:    command.RunE(deprovisonCmd),
	}
	deprovisonCmd.AddNamespaceFlags(cmd.Flags(), false)
	deprovisonCmd.AddWaitFlags(cmd)
	cmd.Flags().BoolVar(
		&deprovisonCmd.abandon,
		"abandon",
		false,
		"Forcefully and immediately delete the resource from Service Catalog ONLY, potentially abandoning any broker resources that you may continue to be charged for.",
	)
	cmd.Flags().BoolVarP(
		&deprovisonCmd.skipPrompt,
		"yes",
		"y",
		false,
		`Automatic yes to prompts. Assume "yes" as answer to all prompts and run non-interactively.`,
	)

	return cmd
}

func (c *deprovisonCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.instanceName = args[0]

	return nil
}

func (c *deprovisonCmd) Run() error {
	return c.deprovision()
}

func (c *deprovisonCmd) deprovision() error {
	var err error
	if c.abandon {
		fmt.Fprintln(c.Output, "This action is not reversible and may cause you to be charged for the broker resources that are abandoned. If you have any bindings for this instance, please delete them manually with svcat unbind --abandon --name bindingName")
		if !c.skipPrompt {
			fmt.Fprintln(c.Output, "Are you sure? [y|n]: ")
			s := bufio.NewScanner(os.Stdin)
			s.Scan()

			err = s.Err()
			if err != nil {
				return err
			}

			if strings.ToLower(s.Text()) != "y" {
				err = fmt.Errorf("aborted abandon operation")
				return err
			}
		}

		// Only delete the instance finalizer here. The bindings will still exist for this instance.
		if err = c.App.RemoveFinalizerForInstance(c.Namespace, c.instanceName); err != nil {
			return err
		}
	}

	err = c.App.Deprovision(c.Namespace, c.instanceName)
	if err != nil {
		return err
	}

	if c.Wait {
		fmt.Fprintln(c.Output, "Waiting for the instance to be deleted...")

		var instance *v1beta1.ServiceInstance
		instance, err = c.App.WaitForInstanceToNotExist(c.Namespace, c.instanceName, c.Interval, c.Timeout)

		// The instance failed to deprovision cleanly, dump out more information on why
		if instance != nil && c.App.IsInstanceFailed(instance) {
			output.WriteInstanceDetails(c.Output, instance)
		}
	}

	if err == nil {
		output.WriteDeletedResourceName(c.Output, c.instanceName)
	}
	return err
}
