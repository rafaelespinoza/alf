package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rafaelespinoza/alf"
)

// fooArgs is named args for the "foo" command.
var fooArgs struct {
	Delta int
	Echo  string
}

const maxDelta = 42

// Foo is a subcommand of the root command. Direct children of the root don't
// need to delegate to child commands. They could also be commands themselves.
var Foo alf.Directive = &alf.Command{
	Description: "a terminal task",
	Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
		name := _Bin + " foo"
		flags := flag.NewFlagSet(name, flag.ExitOnError)
		flags.IntVar(&fooArgs.Delta, "delta", 5, "repeat a string delta times")
		flags.StringVar(&fooArgs.Echo, "echo", "test", "string to repeat")
		flags.Usage = func() {
			fmt.Fprintf(flags.Output(), `Usage:

	%s [flags]

Description:

	Example, repeat a string. Must be <= %d

Flags:

`,
				name, maxDelta)
			flags.PrintDefaults()
		}
		return flags
	},
	Run: func(ctx context.Context) error {
		// The Run function is a good place to perform input validation. This
		// example shows the help menu on invalid data.
		if fooArgs.Delta > maxDelta {
			return fmt.Errorf(
				"delta %d must be <= %d %w",
				fooArgs.Delta, maxDelta, alf.ErrShowUsage,
			)
		}
		for i := 0; i < fooArgs.Delta; i++ {
			fmt.Println(fooArgs.Echo)
		}
		return nil
	},
}
