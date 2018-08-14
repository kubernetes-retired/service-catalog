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
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/pflag"
)

// HasFormatFlags represents a command that can have its output formatted.
type HasFormatFlags interface {
	// ApplyFormatFlags persists the format-related flags:
	// * --output
	ApplyFormatFlags(lags *pflag.FlagSet) error
}

// Formatted is the base command of all svcat commands that support customizable output formats.
type Formatted struct {
	OutputFormat string
}

// NewFormatted command.
func NewFormatted() *Formatted {
	return &Formatted{
		OutputFormat: output.FormatTable,
	}
}

// AddOutputFlags adds common output flags to a command that can have variable output formats.
func (c *Formatted) AddOutputFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&c.OutputFormat, "output", "o", output.FormatTable,
		"The output format to use. Valid options are table, json or yaml. If not present, defaults to table",
	)
}

// ApplyFormatFlags persists the format-related flags:
// * --output
func (c *Formatted) ApplyFormatFlags(flags *pflag.FlagSet) error {
	c.OutputFormat = strings.ToLower(c.OutputFormat)

	switch c.OutputFormat {
	case output.FormatTable, output.FormatJSON, output.FormatYAML:
		return nil
	default:
		return fmt.Errorf("invalid --output format %q, allowed values are: table, json and yaml", c.OutputFormat)
	}
}
