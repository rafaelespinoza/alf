package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

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
		Bravo   bool
		Charlie string
	}

	del := &alf.Delegator{Description: "example delegator with subcommands"}

	// define flags for this parent command.
	parentFlags := flag.NewFlagSet(cmdname, flag.ExitOnError)
	parentFlags.IntVar(&barArgs.Alpha, "alpha", 42, "a number")

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
	del.Flags = parentFlags // share flag data from parent to child command.

	// define subcommands here. The key is the subcommand name.
	del.Subs = map[string]alf.Directive{
		"cities": &alf.Command{
			Description: "print a city name",
			// Setup can be used to generate documentation and to define an
			// independent flag set for the subcommand.
			Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
				name := cmdname + " cities"
				inFlags.Init(name, flag.ExitOnError)
				inFlags.BoolVar(&barArgs.Bravo, "bravo", false, "show a city with a B")
				inFlags.StringVar(&barArgs.Charlie, "charlie", "parker", "customize charlie")

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
				if barArgs.Bravo {
					cities = []string{
						"Bonn, Germany",
						"Balikpapan, Indonesia",
						"Beni Mellal, Morocco",
						"Bello, Colombia",
					}
				} else {
					cities = []string{
						"A Coruña, Spain",
						"Ageo, Japan",
						"Accra, Ghana",
						"Avellaneda, Argentina",
					}
				}
				ind := barArgs.Alpha % len(cities)
				fmt.Printf(
					"city: %q, custom charlie: %q\n",
					cities[ind], barArgs.Charlie,
				)
				return nil
			},
		},
		"oof": &alf.Command{
			Description: "maybe error",
			Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
				name := cmdname + " oof"
				inFlags.Init(name, flag.ExitOnError)
				inFlags.BoolVar(&barArgs.Bravo, "bravo", false, "return an error if true")
				inFlags.StringVar(&barArgs.Charlie, "chuck", "berry", "an alternative charlie")
				inFlags.Usage = func() {
					fmt.Fprintf(inFlags.Output(), `Usage:

	%s %s [flags]

Description:

	Return an error if bravo is true, otherwise be ok.`, _Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					inFlags.PrintDefaults()
				}
				return &inFlags
			},
			Run: func(ctx context.Context) error {
				if barArgs.Bravo {
					return fmt.Errorf("demo force show usage%w", alf.ErrShowUsage)
				}
				fmt.Printf("your alternative charlie %q is %d years old", barArgs.Charlie, barArgs.Alpha)
				return nil
			},
		},
	}

	// This demonstrates a subcommand that is both:
	// - a child of a parent command
	// - and a parent of some child commands
	nested := alf.Delegator{
		Description: "a subcommand (with its own commands) of a subcommand",
		Subs: map[string]alf.Directive{
			"alfa": &alf.Command{
				Description: "terminal command of a nested subcommand",
				Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
					inFlags.Init("nested alfa", flag.ContinueOnError)
					inFlags.Usage = func() { fmt.Println("help for nested.alfa") }
					return &inFlags
				},
				Run: func(ctx context.Context) error {
					fmt.Println("called bar.moar.alfa")
					return nil
				},
			},
			"bravo": &alf.Command{
				Description: "terminal command of a nested subcommand, (returns error)",
				Setup: func(inFlags flag.FlagSet) *flag.FlagSet {
					inFlags.Init("nested bravo", flag.ContinueOnError)
					inFlags.Usage = func() { fmt.Println("help for nested.bravo") }
					return &inFlags
				},
				Run: func(ctx context.Context) error {
					fmt.Println("called bar.moar.bravo")
					return errors.New("demo error")
				},
			},
		},
	}
	nested.Flags = flag.NewFlagSet("nested", flag.ContinueOnError)
	nested.Flags.Usage = func() {
		fmt.Fprintf(nested.Flags.Output(), `Usage:

	%s [flags]

Description:

	Demo of nested Delegator (subcommand has a parent and its own subcommands).

Subcommands:

	%v`, _Bin, strings.Join(nested.DescribeSubcommands(), "\n\t"))
		fmt.Printf("\n\nFlags:\n\n")
		nested.Flags.PrintDefaults()
	}
	del.Subs["nested"] = &nested

	return del
}("bar")
