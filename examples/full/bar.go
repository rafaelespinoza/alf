package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/rafaelespinoza/alf"
)

// Bar is an example Delegator that is a direct child of the root. A Delegator
// will have subcommands of its own and its only job is to hand off control to a
// child command so it can perform a task.
//
// A recommended way to define a command with subcommands (a Delegator) is to
// initialize it within a function. This allows you to share data between the
// parent and child within the function scope.
var Bar = func(cmdname string) alf.Directive {
	// barArgs is named args for the "bar" command.
	var barArgs struct {
		Alpha   int
		Biff    bool
		Charlie string
	}

	del := &alf.Delegator{Description: "example delegator with subcommands"}

	// define flags for this parent command.
	parentFlags := flag.NewFlagSet(cmdname, flag.ExitOnError)
	parentFlags.IntVar(&barArgs.Alpha, "alpha", 42, "a number")
	parentFlags.BoolVar(&barArgs.Biff, "biff", false, "what are ya? chicken?")

	// set up help text.
	parentFlags.Usage = func() {
		fmt.Fprintf(parentFlags.Output(), `Usage:

	%s [flags]

Description:

	Demo of Delegator (parent command with subcommands).

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v`, _Bin, strings.Join(del.DescribeSubcommands(), "\n\t"))
		fmt.Printf("\n\nFlags:\n\n")
		parentFlags.PrintDefaults()
	}
	del.Flags = parentFlags // don't forget this.

	// define subcommands here. The key is the subcommand name.
	del.Subs = map[string]alf.Directive{
		"cities": &alf.Command{
			Description: "print a city name",
			// Setup can be used to generate documentation and to define an
			// independent flag set for the subcommand.
			Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
				name := cmdname + " cities"
				inFlags.Init(name, flag.ExitOnError)
				inFlags.StringVar(&barArgs.Charlie, "chuck", "", "an alternative charlie")

				// help text for subcommand.
				inFlags.Usage = func() {
					fmt.Fprintf(inFlags.Output(), `Usage:

	%s %s [flags]

Description:

	Output a city name. Here are some flags.`, _Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					inFlags.PrintDefaults()
				}
				return &inFlags
			},
			// By now, the flags have been parsed and the subcommand is ready to
			// go. This is also a good place to do input validation.
			Run: func(ctx context.Context) error {
				var cities []string
				if barArgs.Biff {
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
				fmt.Printf("your alternative charlie %q\n", barArgs.Charlie)
				return nil
			},
		},
		"oof": &alf.Command{
			Description: "maybe error",
			Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
				name := cmdname + " oof"
				inFlags.Init(name, flag.ExitOnError)
				inFlags.StringVar(&barArgs.Charlie, "chaz", "", "an alternative charlie")
				inFlags.Usage = func() {
					fmt.Fprintf(inFlags.Output(), `Usage:

	%s %s [flags]

Description:

	Return an error if Biff, otherwise be ok.`, _Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					inFlags.PrintDefaults()
				}
				return &inFlags
			},
			Run: func(ctx context.Context) error {
				if barArgs.Biff {
					return fmt.Errorf("sample error, here's a number %d", barArgs.Alpha)
				}
				fmt.Printf("your alternative charlie %q\n", barArgs.Charlie)
				return nil
			},
		},
	}

	return del
}("bar")
