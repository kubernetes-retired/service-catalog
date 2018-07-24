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

	"errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/pflag"
)

// HasScopedFlags represents a command that can be scoped
// to either namespace level or cluster level resources.
type HasScopedFlags interface {
	// ApplyScopedFlags persists the scope related flags:
	// * --cluster
	ApplyScopedFlags(flags *pflag.FlagSet) error
}

// Force the compiler to check that we are implementing this interface
var _ HasScopedFlags = NewScoped()

// Scoped is the base command of all svcat commands that can be scoped
// to either namespace level or cluster level resources.
type Scoped struct {
	allowAll bool
	rawScope string
	Scope    servicecatalog.Scope
}

// NewScoped creates a new Scoped command.
func NewScoped() *Scoped {
	return &Scoped{}
}

// AddScopedFlags adds the scope-related flags.
// * --scope
func (c *Scoped) AddScopedFlags(flags *pflag.FlagSet, allowAll bool) {
	c.allowAll = allowAll
	if allowAll {
		flags.StringVar(&c.rawScope, "scope", servicecatalog.AllScope, "Limit the results to a particular scope: cluster, namespace or all")
	} else {
		flags.StringVar(&c.rawScope, "scope", servicecatalog.NamespaceScope, "Limit the results to a particular scope: cluster, namespace")
	}
}

// ApplyScopedFlags persists the scope-related flags:
// * --scope
func (c *Scoped) ApplyScopedFlags(flags *pflag.FlagSet) error {
	switch c.rawScope {
	case servicecatalog.AllScope:
		if !c.allowAll {
			return errors.New("invalid --scope (all), allowed values are: cluster, namespace")
		}
		c.Scope = servicecatalog.Scope(c.rawScope)
		return nil
	case servicecatalog.ClusterScope, servicecatalog.NamespaceScope:
		c.Scope = servicecatalog.Scope(c.rawScope)
		return nil
	default:
		return fmt.Errorf("invalid --scope (%s), allowed values are: all, cluster, namespace", c.rawScope)
	}
}
