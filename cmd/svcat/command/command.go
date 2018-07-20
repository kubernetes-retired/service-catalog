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
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

// Command represents an svcat command.
type Command interface {
	// Validate and load the arguments passed to the svcat command.
	Validate(args []string) error

	// Run a validated svcat command.
	Run() error
}

// PreRunE validates os args, and then saves them on the svcat command.
func PreRunE(cmd Command) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if nsCmd, ok := cmd.(HasNamespaceFlags); ok {
			nsCmd.ApplyNamespaceFlags(c.Flags())
		}
		if scopedCmd, ok := cmd.(HasScopedFlags); ok {
			scopedCmd.ApplyScopedFlags(c.Flags())
		}
		if fmtCmd, ok := cmd.(HasFormatFlags); ok {
			err := fmtCmd.ApplyFormatFlags(c.Flags())
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
		if planFilteredCmd, ok := cmd.(HasPlanFlag); ok {
			err := planFilteredCmd.ApplyPlanFlag(c)
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
		// validate the args and print help info if needed.
		err := cmd.Validate(args)
		if err != nil {
			fmt.Println(err)
			fmt.Println(c.UsageString())
		}
		return err
	}
}

// RunE executes a validated svcat command.
func RunE(cmd Command) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		return cmd.Run()
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
