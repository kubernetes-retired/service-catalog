package command

import (
	"io"

	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
)

// Context is ambient data necessary to run any svcat command.
type Context struct {
	// Output should be used instead of directly writing to stdout/stderr, to enable unit testing.
	Output io.Writer

	// svcat application, the library behind the cli
	App *svcat.App
}
