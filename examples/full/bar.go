package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/rafaelespinoza/alf"
)

// BarArgs is named args for the "bar" command.
type BarArgs struct {
	Biff    bool
	Charlie float64
}

// A recommended to define a command with subcommands (a Delegator) is to create
// a function and invoke it right away. This allows you to share data between
// the parent and child within the function scope. The input param, cmdname can
// be used for generating documentation.
var _Bar = func(cmdname string) alf.Directive {
	del := &alf.Delegator{Description: "example delegator with subcommands"}
	var barArgs BarArgs

	// define flags for this parent command.
	flags := flag.NewFlagSet(cmdname, flag.ExitOnError)
	flags.BoolVar(&barArgs.Biff, "biff", false, "bbb")
	flags.Float64Var(&barArgs.Charlie, "charlie", 1.23, "do stuff with floats")

	// set up help text.
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), `Usage:

	%s [flags]

Description:

	Demo of Delegator (parent command with subcommands).

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v`, _Bin, strings.Join(del.DescribeSubcommands(), "\n\t"))
		fmt.Printf("\n\nFlags:\n\n")
		flags.PrintDefaults()
	}
	del.Flags = flags // don't forget this.

	// define subcommands here. The key is the subcommand name.
	del.Subs = map[string]alf.Directive{
		"cities": &alf.Command{
			Description: "print a city name",
			// Setup can be used to generate documentation and to define an
			// independent flag set for the subcommand.
			Setup: func(posArgs []string) *flag.FlagSet {
				name := cmdname + " cities"
				flags := flag.NewFlagSet(name, flag.ExitOnError)
				var barArgs BarArgs

				// help text for subcommand.
				flags.Usage = func() {
					fmt.Fprintf(flags.Output(), `Usage:

	%s %s [flags]

Description:

	Output a city name. Here are some flags.`, _Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				// Remember to assign this, or else Run won't be able to know
				// what's been collected here.
				_Args.Bar = &barArgs
				return flags
			},
			// By now, the flags have been parsed and the subcommand is ready to
			// go. This is also a good place to do input validation.
			Run: func(ctx context.Context, posArgs []string) error {
				args := _Args.Bar
				var cities []string
				if args.Biff {
					cities = []string{"Hill Valley, CA"}
				} else {
					cities = []string{
						"A Coruña, Spain",
						"Ageo, Japan",
						"Accra, Ghana",
						"Avellaneda, Argentina",
					}
				}
				ind := time.Now().Second() % len(cities)
				fmt.Printf("city: %q\n", cities[ind])
				return nil
			},
		},
		"oof": &alf.Command{
			Description: "maybe error",
			Setup: func(posArgs []string) *flag.FlagSet {
				name := cmdname + " oof"
				flags := flag.NewFlagSet(name, flag.ExitOnError)

				flags.Usage = func() {
					fmt.Fprintf(flags.Output(), `Usage:

	%s %s [flags]

Description:

	Return an error if Biff, otherwise be ok.`, _Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				return flags
			},
			Run: func(ctx context.Context, posArgs []string) error {
				args := _Args.Bar
				if args.Biff {
					return fmt.Errorf("sample error, here's a number %f", args.Charlie)
				}
				fmt.Println("ok")
				return nil
			},
		},
	}

	_Args.Bar = &barArgs
	return del
}("bar")
