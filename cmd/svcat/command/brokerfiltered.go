/*
Copyright 2020 The Kubernetes Authors.

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

const (
	// BrokerAll is the default argument to specify when you want to list or filter resources for all brokers
	BrokerAll string = ""
)

// HasBrokerFlag represents a command that supports --broker.
type HasBrokerFlag interface {
	// ApplyBrokerFlag validates and persists the broker related flag.
	//   --broker
	ApplyBrokerFlag(*cobra.Command) error
}

// BrokerFiltered adds support to a command for the --broker flag.
type BrokerFiltered struct {
	BrokerFilter string
}

// NewBrokerFiltered initializes a new broker specified command.
func NewBrokerFiltered() *BrokerFiltered {
	return &BrokerFiltered{}
}

// AddBrokerFlag adds the broker related flag.
//   --broker
func (c *BrokerFiltered) AddBrokerFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"broker",
		"b",
		"",
		"If present, specify the broker used as a filter for this request",
	)
}

// ApplyBrokerFlag persists the broker related flag.
//   --broker
func (c *BrokerFiltered) ApplyBrokerFlag(cmd *cobra.Command) error {
	var err error
	c.BrokerFilter, err = cmd.Flags().GetString("broker")
	return err
}
