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

package versions

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg"
	"github.com/spf13/cobra"
)

type versionCmd struct {
	*command.Context
	client bool
	server bool
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
	cmd.Flags().BoolVarP(
		&versionCmd.client,
		"client",
		"c",
		false,
		"Show only the client version",
	)
	cmd.Flags().BoolVarP(
		&versionCmd.server,
		"server",
		"s",
		false,
		"Show only the server version",
	)

	return cmd
}

func (c *versionCmd) Validate(args []string) error {
	if !c.client && !c.server {
		c.client = true
		c.server = true
	}
	return nil
}

func (c *versionCmd) Run() error {
	return c.version()
}

func (c *versionCmd) version() error {
	client := ""
	if c.client {
		client = pkg.VERSION
	}

	server := ""
	if c.server {
		version, err := c.App.ServerVersion()
		if err != nil {
			return nil
		} else {
			server = version.GitVersion
		}
	}

	output.WriteVersion(c.Output, client, server)
	return nil
}
