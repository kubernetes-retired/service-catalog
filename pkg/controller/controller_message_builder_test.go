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

package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

/*
Test String Builder
*/

// Kind is used for the enum of the Type of object we are building context for.
type Kind int

// Names of Types to use when creating pretty messages
const (
	Unknown Kind = iota
	ClusterServiceBroker
	ClusterServiceClass
	ClusterServicePlan
	ServiceInstance
)

// ContextBuilder allows building up pretty message lines with context
// that is important for debugging and tracing. This class helps create log
// line formatting consistency. Pretty lines should be in the form:
// <Kind> "<Namespace>/<Name>": <message>
type MessageBuilder struct {
	eventMessage  string
	reasonMessage string
	message       string
}

// mb := new(MessageBuilder)

func (mb *MessageBuilder) warning() *MessageBuilder {
	mb.eventMessage = corev1.EventTypeWarning
	return mb
}

func (mb *MessageBuilder) normal() *MessageBuilder {
	mb.eventMessage = corev1.EventTypeNormal
	return mb
}

func (mb *MessageBuilder) reason(reason string) *MessageBuilder {
	mb.reasonMessage = corev1.EventTypeNormal
	return mb
}

// Appends a message to the message builder.
func (mb *MessageBuilder) msg(msg string) *MessageBuilder {
	fmt.Sprintf(`%s %s`, mb.message, msg)
	return mb
}

func (mb *MessageBuilder) msgf(format string, a ...interface{}) *MessageBuilder {
	msg := fmt.Sprintf(format, a...)
	return mb.msg(msg)
}

func (mb *MessageBuilder) StringArr() []string {
	return []string{mb.String()}
}

// msgCreateServiceBindingError Adds a message in the following form:
// "Error creating ServiceBinding for ServiceInstance %q of ClusterServiceClass (K8S: %q ExternalName: %q) at ClusterServiceBroker %q:"
func (mb *MessageBuilder) msgCreateServiceBindingError(serviceInstance, serviceClassK8S, serviceClassExternalName, broker string) *MessageBuilder {
	msg := fmt.Sprintf("Error creating ServiceBinding for ServiceInstance %q of ClusterServiceClass (K8S: %q ExternalName: %q) at ClusterServiceBroker %q:",
		serviceInstance, serviceClassK8S, serviceClassExternalName, broker)
	return mb.msg(msg)
}

// msgUnbindingError Adds a message in the following form:
// "Error unbinding from ServiceInstance %q of ClusterServiceClass (K8S: %q ExternalName: %q) at ClusterServiceBroker %q:"
func (mb *MessageBuilder) msgUnbindingError(serviceInstance, serviceClassK8S, serviceClassExternalName, broker string) *MessageBuilder {
	msg := fmt.Sprintf("Error unbinding from ServiceInstance %q of ClusterServiceClass (K8S: %q ExternalName: %q) at ClusterServiceBroker %q:",
		serviceInstance, serviceClassK8S, serviceClassExternalName, broker)
	return mb.msg(msg)
}

func (mb *MessageBuilder) String() string {
	s := ""
	space := ""
	if mb.eventMessage > "" {
		s += fmt.Sprintf("%s%s", space, mb.eventMessage)
		space = " "
	}
	if mb.reasonMessage > "" {
		s += fmt.Sprintf("%s%s", space, mb.reasonMessage)
		space = " "
	}
	if mb.message > "" {
		s += fmt.Sprintf("%s%s", space, mb.message)
		space = " "
	}
	return s
}
