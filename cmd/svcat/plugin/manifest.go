package plugin

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var reservedFlags = map[string]struct{}{
	"alsologtostderr":       {},
	"as":                    {},
	"as-group":              {},
	"cache-dir":             {},
	"certificate-authority": {},
	"client-certificate":    {},
	"client-key":            {},
	"cluster":               {},
	"context":               {},
	"help":                  {},
	"insecure-skip-tls-verify": {},
	"kubeconfig":               {},
	"kube-context":             {},
	"log-backtrace-at":         {},
	"log-dir":                  {},
	"log-flush-frequency":      {},
	"logtostderr":              {},
	"match-server-version":     {},
	"n":               {},
	"namespace":       {},
	"password":        {},
	"request-timeout": {},
	"s":               {},
	"server":          {},
	"stderrthreshold": {},
	"token":           {},
	"user":            {},
	"username":        {},
	"v":               {},
	"vmodule":         {},
}

type Manifest struct {
	Plugin `yaml:",inline"`
}

type Plugin struct {
	Name      string   `yaml:"name"`
	ShortDesc string   `yaml:"shortDesc"`
	LongDesc  string   `yaml:"longDesc,omitempty"`
	Example   string   `yaml:"example,omitempty"`
	Command   string   `yaml:"command"`
	Flags     []Flag   `yaml:"flags,omitempty"`
	Tree      []Plugin `yaml:"tree,omitempty"`
}

type Flag struct {
	Name      string `yaml:"name"`
	Shorthand string `yaml:"shorthand,omitempty"`
	Desc      string `yaml:"desc"`
	DefValue  string `yaml:"defValue,omitempty"`
}

func (m *Manifest) Load(rootCmd *cobra.Command) {
	m.Plugin = m.convertToPlugin(rootCmd)
}

func (m *Manifest) convertToPlugin(cmd *cobra.Command) Plugin {
	p := Plugin{}

	p.Name = strings.Split(cmd.Use, " ")[0]
	p.ShortDesc = cmd.Short
	if p.ShortDesc == "" {
		p.ShortDesc = " " // The plugin won't validate if empty
	}
	p.LongDesc = cmd.Long
	p.Command = "./" + cmd.CommandPath()

	p.Flags = []Flag{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		result := m.convertToFlag(flag)
		if result != nil {
			p.Flags = append(p.Flags, *result)
		}
	})

	p.Tree = make([]Plugin, len(cmd.Commands()))
	for i, subCmd := range cmd.Commands() {
		p.Tree[i] = m.convertToPlugin(subCmd)
	}
	return p
}

func (m *Manifest) convertToFlag(src *pflag.Flag) *Flag {
	if _, reserved := reservedFlags[src.Name]; reserved {
		return nil
	}

	dest := &Flag{
		Name: src.Name,
		Desc: src.Usage,
	}

	if _, reserved := reservedFlags[src.Shorthand]; !reserved {
		dest.Shorthand = src.Shorthand
	}

	return dest
}
