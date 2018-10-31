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

package extra

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg"
	"github.com/spf13/cobra"
)

// VersionCmd contains the information needed to print the version
// of svcat and the targeted Service Catalog to the user
type VersionCmd struct {
	*command.Context
	Client bool
	Server bool
}

// NewVersionCmd builds a "svcat version" command
func NewVersionCmd(cxt *command.Context) *cobra.Command {
	versionCmd := &VersionCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Provides the version for the Service Catalog client and server",
		Example: command.NormalizeExamples(`
  svcat version
  svcat version --client
`),
		PreRunE: command.PreRunE(versionCmd),
		RunE:    command.RunE(versionCmd),
	}
	cmd.Flags().BoolVarP(
		&versionCmd.Client,
		"client",
		"c",
		false,
		"Show only the client version",
	)

	return cmd
}

// Validate defaults the client and server to true
func (c *VersionCmd) Validate(args []string) error {
	if !c.Client && !c.Server {
		c.Client = true
		c.Server = true
	}
	return nil
}

// Run  determines the versions of svcat and/or
// the platform and prints them
func (c *VersionCmd) Run() error {
	return c.version()
}

func (c *VersionCmd) version() error {
	if c.Client {
		output.WriteClientVersion(c.Output, pkg.VERSION)
	}

	if c.Server {
		version, err := c.App.ServerVersion()
		if err != nil {
			return err
		}
		output.WriteServerVersion(c.Output, version.GitVersion)
	}

	return nil
}
