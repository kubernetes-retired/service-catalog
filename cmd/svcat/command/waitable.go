package command

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// HasWaitFlags represents a command that supports --wait.
type HasWaitFlags interface {
	// ApplyWaitFlags validates and persists the wait related flags.
	//   --wait
	//   --timeout
	//   --interval
	ApplyWaitFlags() error
}

// WaitableCommand adds support to a command for the --wait flags.
type WaitableCommand struct {
	Wait        bool
	rawTimeout  string
	Timeout     *time.Duration
	rawInterval string
	Interval    time.Duration
}

// NewWaitableCommand initializes a new waitable command.
func NewWaitableCommand() *WaitableCommand {
	return &WaitableCommand{}
}

// AddWaitFlags adds the wait related flags.
//   --wait
//   --timeout
//   --interval
func (c *WaitableCommand) AddWaitFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.Wait, "wait", false,
		"Wait until the operation completes.")
	cmd.Flags().StringVar(&c.rawTimeout, "timeout", "5m",
		"Timeout for --wait, specified in human readable format: 30s, 1m, 1h. Specify -1 to wait indefinitely.")
	cmd.Flags().StringVar(&c.rawInterval, "interval", "1s",
		"Poll interval for --wait, specified in human readable format: 30s, 1m, 1h")
}

// ApplyWaitFlags validates and persists the wait related flags.
//   --wait
//   --timeout
//   --interval
func (c *WaitableCommand) ApplyWaitFlags() error {
	if !c.Wait {
		return nil
	}

	if c.rawTimeout != "-1" {
		timeout, err := time.ParseDuration(c.rawTimeout)
		if err != nil {
			return fmt.Errorf("invalid --timeout value (%s)", err)
		}
		c.Timeout = &timeout
	}

	interval, err := time.ParseDuration(c.rawInterval)
	if err != nil {
		return fmt.Errorf("invalid --interval value (%s)", err)
	}
	c.Interval = interval

	return nil
}
