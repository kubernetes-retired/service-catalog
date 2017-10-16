/*
Copyright 2017 The Kubernetes Authors.

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

package pretty

import (
	"fmt"
)

// Kind is used for the enum of the Type of object we are building context for.
type Kind int

const (
	ServiceInstance Kind = 1
)

func (k Kind) String() string {
	switch k {
	case ServiceInstance:
		return "ServiceInstance"
	default:
		return ""
	}
}

// PrettyContextBuilder allows building up pretty message lines with context
// that is important for debugging and tracing. This class helps create log
// line formatting consistency. Pretty lines should be in the form:
// <Kind> "<Namespace>/<Name>": <message>
type PrettyContextBuilder struct {
	Kind      Kind
	Namespace string
	Name      string
}

// NewPrettyContextBuilder returns a new PrettyContextBuilder that can be used to format messages in the
// form `<Kind> "<Namespace>/<Name>": <message>`.
// kind,  namespace, name are all optional.
func NewPrettyContextBuilder(kind Kind, namespace string, name string) *PrettyContextBuilder {
	lb := new(PrettyContextBuilder)
	lb.Kind = kind
	lb.Namespace = namespace
	lb.Name = name
	return lb
}

// SetKind sets the kind to use in the source context for messages.
func (pcb *PrettyContextBuilder) SetKind(k Kind) *PrettyContextBuilder {
	pcb.Kind = k
	return pcb
}

// SetNamespace sets the namespace to use in the source context for messages.
func (pcb *PrettyContextBuilder) SetNamespace(n string) *PrettyContextBuilder {
	pcb.Namespace = n
	return pcb
}

// SetName sets the name to use in the source context for messages.
func (pcb *PrettyContextBuilder) SetName(n string) *PrettyContextBuilder {
	pcb.Name = n
	return pcb
}

// Message returns a string with message prepended with the current source context.
func (pcb *PrettyContextBuilder) Message(msg string) string {
	if pcb.Kind > 0 || pcb.Namespace != "" || pcb.Name != "" {
		return fmt.Sprintf(`%s: %s`, pcb, msg)
	}
	return msg
}

// TODO(n3wscott): Support <type> (K8S: <K8S-Type-Name> ExternalName: <External-Type-Name>)

func (l PrettyContextBuilder) String() string {
	s := ""
	space := ""
	if l.Kind > 0 {
		s += fmt.Sprintf("%s", l.Kind)
		space = " "
	}
	if l.Namespace != "" && l.Name != "" {
		s += fmt.Sprintf(`%s"%s/%s"`, space, l.Namespace, l.Name)
	} else if l.Namespace != "" {
		s += fmt.Sprintf(`%s"%s"`, space, l.Namespace)
	} else if l.Name != "" {
		s += fmt.Sprintf(`%s"%s"`, space, l.Name)
	}
	return s
}
