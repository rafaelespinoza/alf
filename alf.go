package alf

import (
	"context"
	"fmt"
)

// Root is your main, top-level command. Use your program's init function to set
// up its description, flags and subcommands.
type Root struct{ *Delegator }

// Run parses the top-level flags, extracts the positional arguments and
// executes the command. Invoke this from main.
func (r *Root) Run(ctx context.Context, args []string) error {
	r.Flags.Parse(args)
	posArgs := r.Flags.Args()
	var deleg Directive
	err := r.Perform(ctx, &posArgs)
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
		return deleg.Perform(ctx, &posArgs)
	}

	topic := deleg.(*Delegator)
	if err = topic.Perform(ctx, &posArgs); err != nil {
		topic.Flags.Usage()
		return err
	}

	switch subcmd := topic.Selected.(type) {
	case *Command:
		err = subcmd.Perform(ctx, &posArgs)
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
	Perform(ctx context.Context, positionalArgs *[]string) error
}
