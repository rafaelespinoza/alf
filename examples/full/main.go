// Command full is an example implementation of github.com/rafaelespinoza/alf.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rafaelespinoza/alf"
)

var (
	// Root is the parent command for subcommands and their children.
	Root *alf.Root

	// _Bin is the name of the binary file. It's for usage functions.
	_Bin = os.Args[0]
)

func init() {
	const pkg = "github.com/rafaelespinoza/alf"

	// The entry point is just a Delegator that gets embedded in a Root.
	del := &alf.Delegator{
		Description: "demo " + pkg,
		// Associate with subcommands. A Directive could be a something that
		// performs a task (a Command) or something that passes control to its
		// own subcommands (a Delegator).
		Subs: map[string]alf.Directive{
			"foo": Foo,
			"bar": Bar,
		},
		// Build a plain old flag set from the standard library.
		Flags: flag.NewFlagSet("root", flag.ExitOnError),
	}

	// Add a help message.
	del.Flags.Usage = func() {
		fmt.Fprintf(del.Flags.Output(), `Usage:
	%s [flags] subcommand [subflags]

Description:

	%s is a demo of %s.

Subcommands:

	%v

Examples:

	%s [subcommand] -h`,
			_Bin, _Bin, pkg, strings.Join(Root.DescribeSubcommands(), "\n\t"), _Bin)

		fmt.Printf("\n\nFlags:\n\n")
		del.Flags.PrintDefaults()
	}

	// The root command directs you to other delegators and commands.
	Root = &alf.Root{del}
}

func main() {
	// Invoke the (*Root).Run method to execute your top-level command. It
	// parses input flags and goes from there.
	//
	// NOTE: when passing positional arguments, be sure to reslice it from
	// [1:]. At this point, os.Args[0] is the binary itself. Omit this value
	// when starting your application.
	if err := Root.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
