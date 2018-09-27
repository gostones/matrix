package bot

import (
	"fmt"
)

// EchoBot is a simple echo bot
type EchoBot struct{}

func init() {
	RegisterRobot("echo", func() (robot Robot) {
		return new(EchoBot)
	})
}

// Run executes a command
func (b EchoBot) Run(c *Command) string {
	return fmt.Sprintf("%v", c)
}

// Description describes what the robot does
func (b EchoBot) Description() string {
	return "<something>"
}
