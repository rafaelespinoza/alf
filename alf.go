// Package alf is a small, no-frills toolset for creating simple command line
// applications. It's a thin wrapper around the standard library's flag package
// with some stuff to make composition and documentation generation easier. See
// the examples directory for demo usage.
package alf

import (
	"context"
	"errors"
)

// Root is your main, top-level command.
type Root struct {
	*Delegator
	// PrePerform is an optional function to invoke during Run, after the flags
	// have been parsed but before a subcommand is chosen. Run will return early
	// if this function returns an error.
	PrePerform func(ctx context.Context) error
}

// Run parses the top-level flags, extracts the positional arguments and
// executes the command. Invoke this from main with args as os.Args[1:].
func (r *Root) Run(ctx context.Context, args []string) error {
	if err := r.Flags.Parse(args); err != nil {
		return err
	}
	if r.PrePerform != nil {
		err := r.PrePerform(ctx)
		if errors.Is(err, ErrShowUsage) {
			r.Flags.Usage()
		}
		if err != nil {
			return err
		}
	}
	var directive Directive
	err := r.Perform(ctx)
	if r.Selected == nil {
		// either asked for help or asked for unknown command.
		r.Flags.Usage()
	} else {
		directive = r.Selected
	}
	if err != nil {
		return err
	}

	if cmd, ok := directive.(*Command); ok {
		ierr := cmd.Perform(ctx)
		if errors.Is(ierr, ErrShowUsage) {
			cmd.flags.Usage()
		}
		return ierr
	}

	delegator := directive.(*Delegator)
	if err = delegator.Perform(ctx); err != nil {
		delegator.Flags.Usage()
		return err
	}

	subcmd, ok := delegator.Selected.(*Command)
	if !ok {
		// yeah, let's try not to create too deep of a command hierarchy.
		err = errors.New("one does not simply create too many layers of delegators")
	} else {
		err = subcmd.Perform(ctx)
	}
	if errors.Is(err, ErrShowUsage) {
		subcmd.flags.Usage()
	}
	return err
}

// Directive is an abstraction for a parent or child command. A parent would
// delegate to a subcommand, while a subcommand does the actual task.
type Directive interface {
	// Summary provides a short, one-line description.
	Summary() string
	// Perform should either choose a subcommand or do a task.
	Perform(ctx context.Context) error
}

// ErrShowUsage should be returned or wrapped when you want to show the
// command's help menu (run its Usage func) even though the user did not
// specifically request it. It has no text so it doesn't mess up your error
// message if you're use error wrapping.
var ErrShowUsage = errors.New("")
