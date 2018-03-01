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

package touch

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

// NewCmd builds a "svcat touch instance" command
func NewCmd(cxt *command.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "touch",
		Short:   "Make the service catalog attempt to re-provision an instance",
		Example: "svcat touch instance wordpress-mysql-instance",
	}
	cmd.AddCommand(newTouchInstanceCmd(cxt))
	return cmd
}
