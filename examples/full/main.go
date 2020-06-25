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

// Arguments or any other struct is a decent way to define, control, document
// command-line args for your application.
type Arguments struct {
	Foo *FooArgs
	Bar *BarArgs
}

var (
	// _Args is the top-level set of named arguments.
	_Args Arguments
	// _Bin is the name of the binary file.
	_Bin = os.Args[0]
	// _Root is the parent command for subcommands and their children.
	_Root *alf.Root
)

func init() {
	del := &alf.Delegator{
		Description: "demo github.com/rafaelespinoza/alf",
		Subs: map[string]alf.Directive{
			"foo": _Foo,
			"bar": _Bar,
		},
		Flags: flag.CommandLine,
	}

	del.Flags.Usage = func() {
		descriptions := _Root.DescribeSubcommands()
		fmt.Fprintf(_Root.Flags.Output(), `Usage:
	%s [flags] subcommand [subflags]

Description:

	%s is a tool for showing manpages on your system.

Subcommands:

	%v

Examples:

	%s [subcommand] -h
`, _Bin, _Bin, strings.Join(descriptions, "\n\t"), _Bin)
	}

	_Root = &alf.Root{Delegator: del}
}

func main() {
	// Invoke the (*Root).Run method to execute your top-level command. It
	// parses input flags and goes from there.
	//
	// NOTE: when passing positional arguments, be sure to reslice it from
	// [1:]. At this point, os.Args[0] is the binary itself. Omit this value
	// when starting your application.
	if err := _Root.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
