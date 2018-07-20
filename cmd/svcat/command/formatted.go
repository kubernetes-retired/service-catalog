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

	"github.com/spf13/pflag"
)

// FormattedCommand represents a command that can have it's output
// formatted
type FormattedCommand interface {
	// SetFormat sets the commands output format
	SetFormat(format string)
}

// AddOutputFlags adds common output flags to a command that can have variable output formats.
func AddOutputFlags(flags *pflag.FlagSet) {
	flags.StringP(
		"output",
		"o",
		"",
		"The output format to use. Valid options are table, json or yaml. If not present, defaults to table",
	)
}

func determineOutputFormat(flags *pflag.FlagSet) (string, error) {
	format, _ := flags.GetString("output")
	format = strings.ToLower(format)

	switch format {
	case "", "table":
		return "table", nil
	case "json":
		return "json", nil
	case "yaml":
		return "yaml", nil
	default:
		return "", fmt.Errorf("invalid --output format %q, allowed values are table, json and yaml", format)
	}
}
