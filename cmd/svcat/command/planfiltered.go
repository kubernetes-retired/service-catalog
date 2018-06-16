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

package command

import (
	"github.com/spf13/cobra"
)

// HasPlanFlag represents a command that supports --plan.
type HasPlanFlag interface {
	// ApplyPlanFlag validates and persists the plan related flag.
	//   --plan
	ApplyPlanFlag(*cobra.Command) error
}

// PlanFilteredCommand adds support to a command for the --plan flag.
type PlanFilteredCommand struct {
	PlanFilter string
}

// NewPlanFilteredCommand initializes a new plan specified command.
func NewPlanFilteredCommand() *PlanFilteredCommand {
	return &PlanFilteredCommand{}
}

// AddPlanFlag adds the plan related flag.
//   --plan
func (c *PlanFilteredCommand) AddPlanFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"plan",
		"p",
		"",
		"If present, specify the plan used as a filter for this request",
	)
}

// ApplyPlanFlag persists the plan related flag.
//   --plan
func (c *PlanFilteredCommand) ApplyPlanFlag(cmd *cobra.Command) error {
	var err error
	c.PlanFilter, err = cmd.Flags().GetString("plan")
	return err
}
