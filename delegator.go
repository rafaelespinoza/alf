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
	// Subs associates a name with another Directive. The name is what to
	// specify from the command line.
	Subs map[string]Directive
}

// Summary provides a short, one-line description.
func (d *Delegator) Summary() string { return d.Description }

// Perform chooses a subcommand.
func (d *Delegator) Perform(ctx context.Context) error {
	args := d.Flags.Args()
	if len(args) < 1 {
		err := flag.ErrHelp
		maybeCallUsage(err, d.Flags)
		return err
	}

	var err error
	switch first := args[0]; first {
	case "-h", "-help", "--help", "help":
		err = flag.ErrHelp
	default:
		if cmd, ok := d.Subs[first]; !ok {
			err = fmt.Errorf("%w %q", errUnknownCommand, first)
		} else {
			d.Selected = cmd
		}
	}
	if err != nil {
		maybeCallUsage(err, d.Flags)
		return err
	}

	switch selected := d.Selected.(type) {
	case *Command:
		selected.flags = selected.Setup(*d.Flags)
		if err = selected.flags.Parse(args[1:]); err != nil {
			return err
		}
		err = selected.Perform(ctx)
		maybeCallUsage(err, selected.flags)
	case *Delegator:
		f := selected.Flags
		if f == nil {
			return fmt.Errorf("selected Delegator %q requires Flags", args[0])
		}
		if err = f.Parse(args[1:]); err != nil {
			return err
		}
		err = selected.Perform(ctx)
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
