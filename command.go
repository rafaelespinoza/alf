package alf

import (
	"context"
	"flag"
)

// A Command performs a task.
type Command struct {
	// Description should provide a short summary.
	Description string
	// Setup should prepare Args for interpretation by using the pointer to Args
	// with the returned flag set.
	Setup func(a []string) *flag.FlagSet
	// Run is a wrapper function that selects the necessary command line inputs,
	// executes the command and returns any errors.
	Run func(ctx context.Context, args []string) error
}

// Summary provides a short, one-line description.
func (c *Command) Summary() string { return c.Description }

// Perform calls Run to execute the task at hand.
func (c *Command) Perform(ctx context.Context, args *[]string) error { return c.Run(ctx, *args) }
