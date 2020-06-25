package alf

import (
	"context"
	"flag"
)

// A Command performs a task. A parent Delegator passes control to one of these.
type Command struct {
	// Description should provide a short summary.
	Description string
	// Setup is passed the parent flag set and returns a flag set. Call the
	// (*flag.FlagSet).Init method to reuse the parent flags and modify it as
	// needed. You could also just allow the input flagset to pass through. If
	// you don't want to share any flag data between parent and child, then
	// create a new flag set.
	Setup func(parentFlags flag.FlagSet) *flag.FlagSet
	// Run is a wrapper function that selects the necessary command line inputs,
	// executes the command and returns any errors.
	Run func(ctx context.Context, args []string) error
}

// Summary provides a short, one-line description.
func (c *Command) Summary() string { return c.Description }

// Perform calls Run to execute the task at hand.
func (c *Command) Perform(ctx context.Context, args *[]string) error { return c.Run(ctx, *args) }
