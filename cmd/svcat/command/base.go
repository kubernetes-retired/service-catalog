package command

// Base of all svcat commands.
type Base struct {
	*Context
}

// New base command.
func NewBaseCommand(cxt *Context) *Base {
	return &Base{Context: cxt}
}

// GetContext retrieves the command's context.
func (cmd *Base) GetContext() *Context {
	return cmd.Context
}
