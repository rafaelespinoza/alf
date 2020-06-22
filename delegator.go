package alf

import (
	"context"
	"flag"
	"fmt"
	"sort"
)

// A Delegator is a parent to a set of commands. Its sole purpose is to direct
// traffic to a selected command. It can also collect common flag inputs to pass
// on to subcommands.
type Delegator struct {
	// Description should provide a short summary.
	Description string
	// Flags collect and share inputs to its sub directives.
	Flags *flag.FlagSet
	// Selected is the chosen transfer point of control.
	Selected Directive
	// Subs associates a name with a link to another directive. NOTE: one does
	// not simply create too many layers of delegators.
	Subs map[string]Directive
}

// Summary provides a short, one-line description.
func (d *Delegator) Summary() string { return d.Description }

// Perform chooses a subcommand.
func (d *Delegator) Perform(ctx context.Context, a *Arguments) error {
	if len(a.PositionalArgs) < 1 {
		return flag.ErrHelp
	}
	var err error
	switch a.PositionalArgs[0] {
	case "-h", "-help", "--help", "help":
		err = flag.ErrHelp
	default:
		if cmd, ok := d.Subs[a.PositionalArgs[0]]; !ok {
			err = fmt.Errorf("unknown command %q", a.PositionalArgs[0])
		} else {
			d.Selected = cmd
		}
	}
	if err != nil {
		return err
	}

	switch selected := d.Selected.(type) {
	case *Command:
		err = selected.Setup(a).Parse(a.PositionalArgs[1:])
	case *Delegator:
		a.PositionalArgs = a.PositionalArgs[1:] // I, also like to live dangerously
	default:
		err = fmt.Errorf("unsupported value of type %T", selected)
	}
	return err
}

// DescribeSubcommands outputs summaries of each subcommand ordered by name.
func (d *Delegator) DescribeSubcommands() []string {
	descriptions := make([]string, 0)
	for name, subcmd := range d.Subs {
		descriptions = append(
			descriptions,
			fmt.Sprintf("%-20s\t%-40s", name, subcmd.Summary()),
		)
	}
	sort.Strings(descriptions)
	return descriptions
}
