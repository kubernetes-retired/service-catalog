package util

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)


func NewTracedLogger(uid types.UID) *TracedLogger {
	return &TracedLogger{
		header : fmt.Sprintf("[ReqUID: %s ]", uid),
	}
}
type TracedLogger struct {
	header string
}

func (l *TracedLogger) tracedMsgf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s: %s", l.header, msg)
}

func (l *TracedLogger) tracedMsg(args ...interface{}) string {
	msg := fmt.Sprint(args...)
	return fmt.Sprintf("%s: %s", l.header, msg)
}

func (l *TracedLogger) Infof(format string, args ...interface{}) {
	klog.Infof(l.tracedMsgf(format, args...))
}

func (l *TracedLogger) Errorf(format string, args ...interface{}) {
	klog.Errorf(l.tracedMsgf(format, args...))
}

func (l *TracedLogger) Info(args ...interface{}) {
	klog.Infof(l.tracedMsg(args...))
}

func (l *TracedLogger) Error(args ...interface{}) {
	klog.Errorf(l.tracedMsg(args...))
}

