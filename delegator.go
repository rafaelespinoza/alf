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
	// Subs associates a name with another Directive. It's probably best to not
	// create too deep of a hierarchy of Delegators pointing to Delegators. An
	// exception to this recommendation is a Root command with some Delegators
	// as direct childen, which in turn have just one more level of subcommands.
	Subs map[string]Directive
}

// Summary provides a short, one-line description.
func (d *Delegator) Summary() string { return d.Description }

// Perform chooses a subcommand.
func (d *Delegator) Perform(ctx context.Context, pargs *[]string) error {
	if pargs == nil || len(*pargs) < 1 {
		return flag.ErrHelp
	}
	positionalArgs := *pargs

	var err error
	switch first := positionalArgs[0]; first {
	case "-h", "-help", "--help", "help":
		err = flag.ErrHelp
	default:
		if cmd, ok := d.Subs[first]; !ok {
			err = fmt.Errorf("unknown command %q", first)
		} else {
			d.Selected = cmd
		}
	}
	if err != nil {
		return err
	}

	switch selected := d.Selected.(type) {
	case *Command:
		err = selected.Setup(*d.Flags).Parse(positionalArgs[1:])
	case *Delegator:
		// Delegating to another delegator (for example: from the root command
		// to a subcommand), wouldn't work if pargs was a []string (a "value").
		// So we're using a *[]string (a "pointer") instead.
		*pargs = positionalArgs[1:]
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
