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
)

const defaultBoilerPlate = `
# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var (
	completionLong = `
Output shell completion code for the specified shell (bash).
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
	completionShells = map[string]func(cxt *command.Context, cmd *cobra.Command) error{
		"bash": runCompletionBash,
	}
)

// NewCmdCompletion return command for executing "svcat completion" command
func NewCompletionCmd(cxt *command.Context, boilerPlate string) *cobra.Command {
	shells := []string{}
	for s := range completionShells {
		shells = append(shells, s)
	}

	cmd := &cobra.Command{
		Use:     "completion SHELL",
		Short:   "Output shell completion code for the specified shell (bash or zsh).",
		Long:    completionLong,
		Example: completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCompletion(cxt, boilerPlate, cmd, args)
			if err != nil {
				cxt.Output.Write([]byte(err.Error()))
			}
		},
		ValidArgs: shells,
	}

	return cmd
}

// RunCompletion checks given arguments and executes command
func RunCompletion(cxt *command.Context, boilerPlate string, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Shell not specified")
	}
	if len(args) > 1 {
		return fmt.Errorf("Too many arguments. Expected only the shell type")
	}
	run, found := completionShells[args[0]]
	if !found {
		return fmt.Errorf("Unsupported shell type %q", args[0])
	}

	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := cxt.Output.Write([]byte(boilerPlate)); err != nil {
		return err
	}
	return run(cxt, cmd.Parent())
}

func runCompletionBash(cxt *command.Context, kubeadm *cobra.Command) error {
	return kubeadm.GenBashCompletion(cxt.Output)
}
