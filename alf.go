package alf

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// Run parses the flags and executes the command. Invoke this from main.
func Run(ctx context.Context) error {
	flag.Parse()
	_Args.PositionalArgs = flag.Args()
	var deleg Directive
	err := _MainCommand.Perform(ctx, &_Args)
	if _MainCommand.Selected == nil {
		// either asked for help or asked for unknown command.
		flag.Usage()
	} else {
		deleg = _MainCommand.Selected
	}
	if err != nil {
		return err
	}

	if _, ok := deleg.(*Command); ok {
		return deleg.Perform(ctx, &_Args)
	}

	topic := deleg.(*Delegator)
	if err = topic.Perform(ctx, &_Args); err != nil {
		topic.Flags.Usage()
		return err
	}

	switch subcmd := topic.Selected.(type) {
	case *Command:
		err = subcmd.Perform(ctx, &_Args)
	case *Delegator:
		err = fmt.Errorf("too much delegation, selected should be a %T", &Command{})
	default:
		err = fmt.Errorf("unhandled type %T", subcmd)
	}
	return err
}

var (
	// _Args is a shared top-level arguments value.
	_Args Arguments
	// _Bin is the name of the binary file.
	_Bin = os.Args[0]
	// _MainCommand is the parent command for subcommands and their children.
	_MainCommand *Delegator
)

// Directive is an abstraction for a parent or child command. A parent would
// delegate to a subcommand, while a subcommand does the actual task.
type Directive interface {
	// Summary provides a short, one-line description.
	Summary() string
	// Perform should either choose a subcommand or do a task.
	Perform(ctx context.Context, a *Arguments) error
}

type Arguments struct {
	PositionalArgs []string
}
