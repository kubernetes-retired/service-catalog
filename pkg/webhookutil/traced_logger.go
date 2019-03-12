/*
Copyright 2019 The Kubernetes Authors.

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

package webhookutil

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// TracedLogger is a helper wrapper around the klog to ensure that
// given UID is always logged.
// Is workaround cause in klog we do not have options to create a logger with Field
// which will be added for each logged entry like we can do in logrus or zap.
type TracedLogger struct {
	header string
}

// NewTracedLogger returns new instance of the TracedLogger
func NewTracedLogger(uid types.UID) *TracedLogger {
	return &TracedLogger{
		header: fmt.Sprintf("[ReqUID: %s ]", uid),
	}
}

// Infof logs to the INFO log.
func (l *TracedLogger) Infof(format string, args ...interface{}) {
	klog.Infof(l.tracedMsgf(format, args...))
}

// Errorf logs to the ERROR, WARNING, and INFO logs.
func (l *TracedLogger) Errorf(format string, args ...interface{}) {
	klog.Errorf(l.tracedMsgf(format, args...))
}

// Info logs to the INFO log.
func (l *TracedLogger) Info(args ...interface{}) {
	klog.Info(l.tracedMsg(args...))
}

// Error logs to the ERROR, WARNING, and INFO logs.
func (l *TracedLogger) Error(args ...interface{}) {
	klog.Error(l.tracedMsg(args...))
}

func (l *TracedLogger) tracedMsgf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s: %s", l.header, msg)
}

func (l *TracedLogger) tracedMsg(args ...interface{}) string {
	msg := fmt.Sprint(args...)
	return fmt.Sprintf("%s: %s", l.header, msg)
}
