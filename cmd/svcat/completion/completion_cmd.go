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
	"bytes"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
	"io"
)

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

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	zshInitialization := `
__svcat_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__svcat_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__svcat_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__svcat_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__svcat_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__svcat_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__svcat_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__svcat_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__svcat_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__svcat_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__svcat_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
		printf %q "$1"
	fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi

__svcat_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__svcat_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__svcat_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__svcat_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__svcat_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__svcat_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__svcat_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__svcat_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	cmd.Root().GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__svcat_bash_source <(__svcat_convert_bash_to_zsh)
`
	out.Write([]byte(zshTail))
	return nil
}
