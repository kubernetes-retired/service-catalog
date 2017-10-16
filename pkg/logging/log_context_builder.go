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

package logging

import (
	"fmt"
)

// LogContextBuilder allows building up log lines with context that is important
// for debugging and tracing. This class helps create log line formatting
// consistantly. Logging should always be in the form:
// <kind> "<Namespace>/<Name>": <msg>
type LogContextBuilder struct {
	Kind      string
	Namespace string
	Name      string
}

type Kind int

const (
	ServiceInstance Kind = iota
)

func (k Kind) String() string {
	switch k {
	case ServiceInstance:
		return "ServiceInstance"
	default:
		return ""
	}
}

func NewLogContextBuilder(kind Kind, namespace string, name string) *LogContextBuilder {
	lb := new(LogContextBuilder)
	lb.Kind = kind.String()
	lb.Namespace = namespace
	lb.Name = name
	return lb
}

func (l *LogContextBuilder) SetKind(k string) *LogContextBuilder {
	l.Kind = k
	return l
}

func (l *LogContextBuilder) SetNamespace(n string) *LogContextBuilder {
	l.Namespace = n
	return l
}

func (l *LogContextBuilder) SetName(n string) *LogContextBuilder {
	l.Name = n
	return l
}

func (l *LogContextBuilder) Message(msg string) string {
	if l.Kind != "" || l.Namespace != "" || l.Name != "" {
		return fmt.Sprintf(`%s: %s`, l, msg)
	}
	return msg
}

// TODO(n3wscott): Support <type> (K8S: <K8S-Type-Name> ExternalName: <External-Type-Name>)

func (l LogContextBuilder) String() string {
	s := ""
	space := ""
	if l.Kind != "" {
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
