/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"fmt"
	"github.com/spf13/pflag"
	"strings"
)

const (
	removeCRD = "remove-crd"

	webhookConfigurationsNames       = "webhook-configurations"
	serviceCatalogNamespaceParameter = "service-catalog-namespace"
	controllerManagerNameParameter   = "controller-manager-deployment"
)

// CleanerOptions holds configuration for cleaner jobs
type CleanerOptions struct {
	Command               string
	WebhookConfigurations string
	ReleaseNamespace      string
	ControllerManagerName string
}

// NewCleanerOptions creates and returns a new CleanerOptions
func NewCleanerOptions() *CleanerOptions {
	return &CleanerOptions{}
}

// AddFlags adds flags for a CleanerOptions to the specified FlagSet.
func (c *CleanerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Command, "cleaner-command", removeCRD, "Command name to execute")

	fs.StringVar(&c.WebhookConfigurations, webhookConfigurationsNames, "", "Names of Webhook Configurations")
	fs.StringVar(&c.ReleaseNamespace, serviceCatalogNamespaceParameter, "", "Name of namespace where Service Catalog is released")
	fs.StringVar(&c.ControllerManagerName, controllerManagerNameParameter, "", "Name of controller manager deployment")
}

// Validate checks flag has been set and has a proper value
func (c *CleanerOptions) Validate() error {
	switch c.Command {
	case removeCRD:
		return checkParameters(removeCRD, map[string]string{
			webhookConfigurationsNames:       c.WebhookConfigurations,
			serviceCatalogNamespaceParameter: c.ReleaseNamespace,
			controllerManagerNameParameter:   c.ControllerManagerName,
		})
	default:
		return fmt.Errorf("Command %q is not supported", c.Command)
	}
}

// WebhookConfigurationsName returns webhook configuration names based on flags passed to application
func (c *CleanerOptions) WebhookConfigurationsName() []string {
	return strings.Split(c.WebhookConfigurations, " ")
}

func checkParameters(cmd string, params map[string]string) error {
	for name, value := range params {
		if value == "" {
			return fmt.Errorf("command %q requires %d parameter(s), parameter (%q) is empty", cmd, len(params), name)
		}
	}

	return nil
}
