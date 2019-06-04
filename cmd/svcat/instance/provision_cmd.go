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
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/parameters"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// ProvisionCmd contains the info needed to provision a new service instance
type ProvisionCmd struct {
	*command.Namespaced
	*command.Waitable

	ClassKubeName            string
	ClassName                string
	ExternalID               string
	InstanceName             string
	JSONParams               string
	LookupByKubeName         bool
	Params                   interface{}
	PlanKubeName             string
	PlanName                 string
	ProvisionClusterInstance bool
	RawParams                []string
	RawSecrets               []string
	Secrets                  map[string]string
}

// NewProvisionCmd builds a "svcat provision" command
func NewProvisionCmd(cxt *command.Context) *cobra.Command {
	provisionCmd := &ProvisionCmd{
		Namespaced: command.NewNamespaced(cxt),
		Waitable:   command.NewWaitable(),
	}
	cmd := &cobra.Command{
		Use:   "provision NAME --plan PLAN --class CLASS",
		Short: "Create a new instance of a service",
		Example: command.NormalizeExamples(`
  svcat provision wordpress-mysql-instance --class mysqldb --plan free -p location=eastus -p sslEnforcement=disabled
  svcat provision wordpress-mysql-instance --external-id a7c00676-4398-11e8-842f-0ed5f89f718b --class mysqldb --plan free
  svcat provision wordpress-mysql-instance --class mysqldb --plan free -s mysecret[dbparams]
  svcat provision secure-instance --class mysqldb --plan secureDB --params-json '{
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
		PreRunE: command.PreRunE(provisionCmd),
		RunE:    command.RunE(provisionCmd),
	}
	cmd.Flags().StringVar(&provisionCmd.ClassName, "class", "", "The class name (Required)")
	cmd.MarkFlagRequired("class")
	cmd.Flags().StringVar(&provisionCmd.PlanName, "plan", "", "The plan name (Required)")
	cmd.MarkFlagRequired("plan")
	cmd.Flags().StringVar(&provisionCmd.ExternalID, "external-id", "", "The ID of the instance for use with the OSB SB API (Optional)")
	cmd.Flags().BoolVarP(&provisionCmd.LookupByKubeName, "kube-name", "k", false, "Whether or not to interpret the Class/Plan names as Kubernetes names (the default is by external name)")
	cmd.Flags().StringSliceVarP(&provisionCmd.RawParams, "param", "p", nil, "Additional parameter to use when provisioning the service, format: NAME=VALUE. Cannot be combined with --params-json, Sensitive information should be placed in a secret and specified with --secret")
	cmd.Flags().StringVar(&provisionCmd.JSONParams, "params-json", "", "Additional parameters to use when provisioning the service, provided as a JSON object. Cannot be combined with --param")
	cmd.Flags().StringSliceVarP(&provisionCmd.RawSecrets, "secret", "s", nil, "Additional parameter, whose value is stored in a secret, to use when provisioning the service, format: SECRET[KEY]")
	provisionCmd.AddNamespaceFlags(cmd.Flags(), false)
	provisionCmd.AddWaitFlags(cmd)

	return cmd
}

// Validate ensures the required args were provided
// and parses provided params and secrets
func (c *ProvisionCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.InstanceName = args[0]

	var err error

	if c.JSONParams != "" && len(c.RawParams) > 0 {
		return fmt.Errorf("--params-json cannot be used with --param")
	}

	if c.JSONParams != "" {
		c.Params, err = parameters.ParseVariableJSON(c.JSONParams)
		if err != nil {
			return fmt.Errorf("invalid --params-json value (%s)", err)
		}
	} else {
		c.Params, err = parameters.ParseVariableAssignments(c.RawParams)
		if err != nil {
			return fmt.Errorf("invalid --param value (%s)", err)
		}
	}

	c.Secrets, err = parameters.ParseKeyMaps(c.RawSecrets)
	if err != nil {
		return fmt.Errorf("invalid --secret value (%s)", err)
	}

	return nil
}

// Run calls the Provision method
func (c *ProvisionCmd) Run() error {
	err := c.findKubeNames()
	if err != nil {
		return err
	}
	return c.provision()
}

// FindKubeNames determines if we need to find the Kubernetes
// metadata names of the Class/Plan, and finds them if we do.
// It also sets whether we are provisioning a ClusterServiceClass
// or ServiceClass instance
func (c *ProvisionCmd) findKubeNames() error {
	scopeOpts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     servicecatalog.AllScope,
	}
	if c.LookupByKubeName {
		c.ClassKubeName = c.ClassName
		c.PlanKubeName = c.PlanName

		class, err := c.App.RetrieveClassByID(c.ClassKubeName, scopeOpts)
		if err != nil {
			return err
		}
		c.ProvisionClusterInstance = class.IsClusterServiceClass()
		return nil
	} // else lookup by external name
	class, err := c.App.RetrieveClassByName(c.ClassName, scopeOpts)
	if err != nil {
		if strings.Contains(err.Error(), "more than one matching class") {
			return fmt.Errorf("More than one class '%s' found, please specify Kubernetes names using --kube-name", c.ClassName)
		}
		return err
	}
	c.ClassKubeName = class.GetName()
	c.ProvisionClusterInstance = class.IsClusterServiceClass()
	if class.IsClusterServiceClass() {
		scopeOpts.Scope = servicecatalog.ClusterScope
	} else {
		scopeOpts.Scope = servicecatalog.NamespaceScope
	}
	plan, err := c.App.RetrievePlanByClassIDAndName(c.ClassKubeName, c.PlanName, scopeOpts)
	if err != nil {
		return fmt.Errorf("Unable to find plan '%s': %s", c.PlanName, err.Error())
	}
	c.PlanKubeName = plan.GetName()
	return nil
}

// Provision calls the pkg/svcat lib to provision the instance,
// waits if necessary, and then displays the created instance
// to the user
func (c *ProvisionCmd) provision() error {
	opts := &servicecatalog.ProvisionOptions{
		ExternalID: c.ExternalID,
		Namespace:  c.Namespace,
		Params:     c.Params,
		Secrets:    c.Secrets,
	}
	instance, err := c.App.Provision(c.InstanceName, c.ClassKubeName, c.PlanKubeName, c.ProvisionClusterInstance, opts)
	if err != nil {
		return err
	}

	if c.Wait {
		fmt.Fprintln(c.Output, "Waiting for the instance to be provisioned...")
		finalInstance, err := c.App.WaitForInstance(instance.Namespace, instance.Name, c.Interval, c.Timeout)
		if err == nil {
			instance = finalInstance
		}

		// Always print the instance because the provision did succeed,
		// and just print any errors that occurred while polling
		output.WriteInstanceDetails(c.Output, instance)
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)
	return nil
}
