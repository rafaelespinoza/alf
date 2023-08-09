package alf

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
)

// A Delegator is a parent to a set of commands. Its sole purpose is to direct
// traffic to a selected command. It can also collect common flag inputs to pass
// on to subcommands.
type Delegator struct {
	// Description should provide a short summary.
	Description string
	// Flags collect and share inputs to its sub directives.
	Flags *flag.FlagSet
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

	var (
		name     string
		err      error
		selected Directive
	)
	switch name = args[0]; name {
	case "-h", "-help", "--help", "help":
		err = flag.ErrHelp
	default:
		cmd, found := d.chooseSubcommand(name)
		if !found {
			err = fmt.Errorf("%w %q", errUnknownCommand, name)
		} else if cmd == nil {
			err = fmt.Errorf("subcommand %q is empty", name)
		} else {
			selected = cmd
		}
	}
	if err != nil {
		maybeCallUsage(err, d.Flags)
		return err
	}

	switch selected := selected.(type) {
	case *Command:
		if selected.Setup == nil {
			return fmt.Errorf("subcommand %q Setup is empty", name)
		}
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

func (d *Delegator) chooseSubcommand(name string) (out Directive, found bool) {
	out, found = d.Subs[name]
	if found {
		return
	}
	if name == "" {
		return
	}

	for fullname, directive := range d.Subs {
		if strings.HasPrefix(fullname, name) {
			out = directive
			found = true
			return
		}
	}

	return
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
