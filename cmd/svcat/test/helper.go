package svcattest

import (
	"io"
	"io/ioutil"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/spf13/viper"
)

// NewContext creates a test context for the svcat cli, optionally capturing the
// command output, or injecting a fake set of clients.
func NewContext(outputCapture io.Writer, fakeApp *svcat.App) *command.Context {
	if outputCapture == nil {
		outputCapture = ioutil.Discard
	}

	return &command.Context{
		Viper:  viper.New(),
		Output: outputCapture,
		App:    fakeApp,
	}
}
