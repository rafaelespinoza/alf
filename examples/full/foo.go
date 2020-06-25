package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rafaelespinoza/alf"
)

// FooArgs is named args for the "foo" command.
type FooArgs struct {
	Delta int
	Echo  string
}

var _Foo = &alf.Command{
	Description: "a terminal task",
	Setup: func(posArgs []string) *flag.FlagSet {
		name := _Bin + " foo"
		flags := flag.NewFlagSet(name, flag.ExitOnError)
		var fooArgs FooArgs
		flags.IntVar(&fooArgs.Delta, "delta", 5, "repeat a string delta times")
		flags.StringVar(&fooArgs.Echo, "echo", "test", "string to repeat")
		flags.Usage = func() {
			fmt.Fprintf(flags.Output(), `Usage:

	%s [flags]

Description:

	Example, repeat a string.`, _Bin)
			fmt.Printf("\n\nFlags:\n\n")
			flags.PrintDefaults()
		}
		_Args.Foo = &fooArgs
		return flags
	},
	Run: func(ctx context.Context, posArgs []string) error {
		args := _Args.Foo
		for i := 0; i < args.Delta; i++ {
			fmt.Println(args.Echo)
		}
		return nil
	},
}
