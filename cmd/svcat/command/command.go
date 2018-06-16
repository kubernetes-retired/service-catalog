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
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command represents an svcat command.
type Command interface {
	// Validate and load the arguments passed to the svcat command.
	Validate(args []string) error

	// Run a validated svcat command.
	Run() error
}

// NamespacedCommand represents a command that can be scoped to a namespace.
type NamespacedCommand interface {
	Command

	// GetContext retrieves the command's context.
	GetContext() *Context

	// SetNamespace sets the effective namespace for the command.
	SetNamespace(namespace string)
}

// FormattedCommand represents a command that can have it's output
// formatted
type FormattedCommand interface {
	// SetFormat sets the commands output format
	SetFormat(format string)
}

// PreRunE validates os args, and then saves them on the svcat command.
func PreRunE(cmd Command) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if nsCmd, ok := cmd.(NamespacedCommand); ok {
			namespace := DetermineNamespace(c.Flags(), nsCmd.GetContext().App.CurrentNamespace)
			nsCmd.SetNamespace(namespace)
		}
		if fmtCmd, ok := cmd.(FormattedCommand); ok {
			fmtString, err := determineOutputFormat(c.Flags())
			if err != nil {
				return err
			}
			fmtCmd.SetFormat(fmtString)
		}
		if planFilteredCmd, ok := cmd.(HasPlanFlag); ok {
			err := planFilteredCmd.ApplyPlanFlag(c)
			if err != nil {
				return err
			}
		}
		if classFilteredCmd, ok := cmd.(HasClassFlag); ok {
			err := classFilteredCmd.ApplyClassFlag(c)
			if err != nil {
				return err
			}
		}
		if waitCmd, ok := cmd.(HasWaitFlags); ok {
			err := waitCmd.ApplyWaitFlags()
			if err != nil {
				return err
			}
		}
		return cmd.Validate(args)
	}
}

// RunE executes a validated svcat command.
func RunE(cmd Command) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		return cmd.Run()
	}
}

// AddNamespaceFlags applies the --namespace and --all-namespaces flags to a command.
// This is intended to be used in conjunction with the NamespacedCommand interface.
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this request",
	)

	if allowAll {
		flags.Bool(
			"all-namespaces",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace",
		)
	}
}

// DetermineNamespace using the current context's namespace, and the user-requested namespace.
func DetermineNamespace(flags *pflag.FlagSet, currentNamespace string) string {
	namespace, _ := flags.GetString("namespace")
	allNamespaces, _ := flags.GetBool("all-namespaces")

	if allNamespaces {
		return ""
	}

	if namespace != "" {
		return namespace
	}

	return currentNamespace
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

// NormalizeExamples removes leading and trailing empty lines
// from the command's Example string and normalizes the indentation
// so that all examples across all commands are indented consistently.
func NormalizeExamples(examples string) string {
	// TODO: this code copied from a pending PR: https://github.com/kubernetes/kubernetes/pull/64017; replace this with a call to that method when PR is merged
	indentedLines := []string{}
	var baseIndentation *string
	for _, line := range strings.Split(examples, "\n") {
		if baseIndentation == nil {
			if len(strings.TrimSpace(line)) == 0 {
				continue // skip initial lines that only contain whitespace
			}
			whitespaceAtFront := line[:strings.Index(line, strings.TrimSpace(line))]
			baseIndentation = &whitespaceAtFront
		}
		trimmed := strings.TrimPrefix(line, *baseIndentation)
		indented := "  " + trimmed
		indentedLines = append(indentedLines, indented)
	}
	indentedString := strings.Join(indentedLines, "\n")
	return strings.TrimRightFunc(indentedString, unicode.IsSpace)
}
