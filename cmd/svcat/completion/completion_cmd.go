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

package completion

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
	"io"
)

var (
	completionLong = `
Output shell completion code for the specified shell (bash or zsh).
The shell code must be evaluated to provide interactive
completion of svcat commands. This can be done by sourcing it from
the .bash_profile.

Note: this requires the bash-completion framework, which is not installed
by default on Mac. This can be installed by using homebrew:

	$ brew install bash-completion

Once installed, bash_completion must be evaluated. This can be done by adding the
following line to the .bash_profile

	$ source $(brew --prefix)/etc/bash_completion
`

	completionExample = `
# Install bash completion on a Mac using homebrew
brew install bash-completion
printf "\n# Bash completion support\nsource $(brew --prefix)/etc/bash_completion\n" >> $HOME/.bash_profile
source $HOME/.bash_profile

# Load the svcat completion code for bash into the current shell
source <(svcat completion bash)

# Write bash completion code to a file and source if from .bash_profile
svcat completion bash > ~/.svcat/svcat_completion.bash.inc
printf "\n# Svcat shell completion\nsource '$HOME/.svcat/svcat_completion.bash.inc'\n" >> $HOME/.bash_profile
source $HOME/.bash_profile
`
)

var (
	completionShells = map[string]func(w io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
		"zsh":  runCompletionZsh,
	}
)

type completionCmd struct {
	*command.Context
	command  *cobra.Command
	shellgen func(w io.Writer, cmd *cobra.Command) error
}

// NewCompletionCmd return command for executing "svcat completion" command
func NewCompletionCmd(cxt *command.Context) *cobra.Command {
	completionCmd := &completionCmd{Context: cxt}

	shells := []string{}
	for s := range completionShells {
		shells = append(shells, s)
	}

	cmd := &cobra.Command{
		Use:       "completion SHELL",
		Short:     "Output shell completion code for the specified shell (bash or zsh).",
		Long:      completionLong,
		Example:   completionExample,
		PreRunE:   command.PreRunE(completionCmd),
		RunE:      command.RunE(completionCmd),
		ValidArgs: shells,
	}

	completionCmd.command = cmd

	return cmd
}

func (c *completionCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Shell not specified")
	}
	if len(args) > 1 {
		return fmt.Errorf("Too many arguments. Expected only the shell type")
	}
	gen, found := completionShells[args[0]]
	if !found {
		return fmt.Errorf("Unsupported shell type %q", args[0])
	}

	c.shellgen = gen

	return nil
}

func (c *completionCmd) Run() error {
	return c.shellgen(c.Output, c.command)
}

func runCompletionBash(w io.Writer, cmd *cobra.Command) error {
	return cmd.Root().GenBashCompletion(w)
}

func runCompletionZsh(w io.Writer, cmd *cobra.Command) error {
	return cmd.Root().GenZshCompletion(w)
}
