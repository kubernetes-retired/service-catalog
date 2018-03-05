package command

// Namespaced is the base command of all svcat commands that are namespace scoped.
type Namespaced struct {
	*Base
	Namespace string
}

// New namespaced command.
func NewNamespacedCommand(cxt *Context) *Namespaced {
	return &Namespaced{Base: NewBaseCommand(cxt)}
}

// SetNamespace sets the effective namespace for the command.
func (c *Namespaced) SetNamespace(namespace string) {
	c.Namespace = namespace
}
