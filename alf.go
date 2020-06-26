// Package alf is a small, no-frills toolset for creating simple command line
// applications. It's a thin wrapper around the standard library's flag package
// with some stuff to make composition and documentation generation easier. See
// the examples directory for demo usage.
package alf

import (
	"context"
	"fmt"
)

// Root is your main, top-level command.
type Root struct{ *Delegator }

// Run parses the top-level flags, extracts the positional arguments and
// executes the command. Invoke this from main.
func (r *Root) Run(ctx context.Context, args []string) error {
	if err := r.Flags.Parse(args); err != nil {
		return err
	}
	r.positionalArgs = args
	var deleg Directive
	err := r.Perform(ctx)
	if r.Selected == nil {
		// either asked for help or asked for unknown command.
		r.Flags.Usage()
	} else {
		deleg = r.Selected
	}
	if err != nil {
		return err
	}

	if _, ok := deleg.(*Command); ok {
		return deleg.Perform(ctx)
	}

	topic := deleg.(*Delegator)
	if err = topic.Perform(ctx); err != nil {
		topic.Flags.Usage()
		return err
	}

	switch subcmd := topic.Selected.(type) {
	case *Command:
		err = subcmd.Perform(ctx)
	case *Delegator:
		err = fmt.Errorf("too much delegation, selected should be a %T", &Command{})
	default:
		err = fmt.Errorf("unhandled type %T", subcmd)
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
