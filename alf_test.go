package alf_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/rafaelespinoza/alf"
)

var errStub = errors.New("oof")

func newStubRoot(name string, usage *string) alf.Root {
	var (
		foo string
		bar int
		qux bool
	)
	firstUsage := func(val string) {
		if usage == nil || *usage == "" {
			*usage = val
		}
	}

	delta := alf.Delegator{
		Description: "subcmd with more subs",
		Flags:       newMutedFlagSet("delta", flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"echo": &alf.Command{
				Description: "decorate (mutate) the input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					p.Init("delta echo", flag.ContinueOnError)
					p.BoolVar(&qux, "quebec", true, "qqq")
					p.Usage = func() { firstUsage("root.delta.echo") }
					return &p
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"foxtrot": &alf.Command{
				Description: "ignore input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("delta foxtrot", flag.ContinueOnError)
					f.Usage = func() { firstUsage("root.delta.foxtrot") }
					f.BoolVar(&qux, "qux", false, "QQQ")
					return f
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"golf": &alf.Command{
				Description: "allow input flags to pass through",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					return &p
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"hotel": &alf.Command{
				Description: "force show help directive command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("delta hotel", flag.ContinueOnError)
					f.Usage = func() { firstUsage("root.delta.hotel") }
					return f
				},
				Run: func(ctx context.Context) error { return alf.ErrShowUsage },
			},
		},
	}
	delta.Flags.IntVar(&bar, "bar", 2, "bbb")
	delta.Flags.Usage = func() { firstUsage("root.delta") }

	root := alf.Delegator{
		Description: "root",
		Flags:       newMutedFlagSet(name, flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"alpha": &alf.Command{
				Description: "ok command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("alpha", flag.ContinueOnError)
					f.Usage = func() { firstUsage("root.alpha") }
					return f
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"bravo": &alf.Command{
				Description: "error command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("bravo", flag.ContinueOnError)
					f.Usage = func() { firstUsage("root.bravo") }
					return f
				},
				Run: func(ctx context.Context) error { return errStub },
			},
			"charlie": &alf.Command{
				Description: "force show help command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("charlie", flag.ContinueOnError)
					f.Usage = func() { firstUsage("root.charlie") }
					return f
				},
				Run: func(ctx context.Context) error { return alf.ErrShowUsage },
			},
			"delta": &delta,
		},
	}
	root.Flags.StringVar(&foo, "foo", "golf", "fff")
	root.Flags.Usage = func() { firstUsage("root") }

	return alf.Root{Delegator: &root}
}

// newMutedFlagSet creates a flag set that doesn't output anything when its
// Usage function is invoked. It's kind of annoying to run tests in verbose mode
// but have all this `Usage of foo:` all over the place.
func newMutedFlagSet(name string, exit flag.ErrorHandling) *flag.FlagSet {
	flags := flag.NewFlagSet(name, exit)
	flags.SetOutput(bytes.NewBuffer(nil))
	return flags
}

func TestRoot(t *testing.T) {
	var prePerformTest string
	tests := []struct {
		args          []string
		prePerform    func(context.Context) error
		expPrePerform string
		expErr        bool
		expUsage      string
	}{
		// Help
		{args: []string{""}, expErr: true, expUsage: "root"},
		{args: []string{"-h"}, expErr: true, expUsage: "root"},
		{args: []string{"--help"}, expErr: true, expUsage: "root"},
		// Returns whatever the Command returns.
		{args: []string{"alpha"}, expErr: false},
		{args: []string{"bravo"}, expErr: true},
		// Help on Command
		{args: []string{"charlie"}, expErr: true, expUsage: "root.charlie"},
		// Works with flag values on Root's FlagSet.
		{args: []string{"-foo", "fff", "alpha"}, expErr: false},
		// Parent flags should be specified before the subcommand.
		{args: []string{"alpha", "-foo", "fff"}, expErr: true, expUsage: "root.alpha"},
		// Help on a Delegator.
		{args: []string{"delta"}, expErr: true, expUsage: "root.delta"},
		{args: []string{"delta", "-h"}, expErr: true, expUsage: "root.delta"},
		// Access Delegator -> Command.
		{args: []string{"delta", "echo"}, expErr: false},
		{args: []string{"delta", "foxtrot"}, expErr: false},
		{args: []string{"delta", "golf"}, expErr: false},
		// Help on Delegator -> Command.
		{args: []string{"delta", "echo", "-h"}, expErr: true, expUsage: "root.delta.echo"},
		{args: []string{"delta", "foxtrot", "-h"}, expErr: true, expUsage: "root.delta.foxtrot"},
		{args: []string{"delta", "golf", "-h"}, expErr: true, expUsage: "root.delta"},
		{args: []string{"delta", "hotel"}, expErr: true, expUsage: "root.delta.hotel"},
		// Unknown command.
		{args: []string{"echo"}, expErr: true, expUsage: "root"},
		{args: []string{"delta", "india"}, expErr: true, expUsage: "root.delta"},
		// Optional PrePerform field.
		{
			args: []string{"alpha"},
			prePerform: func(ctx context.Context) error {
				prePerformTest = "alpha prePerformCheck"
				return nil
			},
			expErr:        false,
			expPrePerform: "alpha prePerformCheck",
		},
		// This args sequence would normally not have an error.
		{
			args:       []string{"delta", "foxtrot"},
			prePerform: func(ctx context.Context) error { return errStub },
			expErr:     true,
		},
		{
			args:       []string{"delta", "foxtrot"},
			prePerform: func(ctx context.Context) error { return alf.ErrShowUsage },
			expErr:     true,
			expUsage:   "root",
		},
		{
			args: []string{"delta", "foxtrot"},
			prePerform: func(ctx context.Context) error {
				return fmt.Errorf("%w, wrap", alf.ErrShowUsage)
			},
			expErr:   true,
			expUsage: "root",
		},
		// Shouldn't get in the way of parent flags.
		{
			args:       []string{"-foo", "fff", "alpha"},
			prePerform: func(ctx context.Context) error { return nil },
			expErr:     false,
		},
	}

	for i, test := range tests {
		prePerformTest = ""
		var usage string
		root := newStubRoot(fmt.Sprintf("%s %d", t.Name(), i), &usage)
		if test.prePerform != nil {
			root.PrePerform = test.prePerform
		}

		err := root.Run(context.TODO(), test.args)
		if err == nil && test.expErr {
			t.Errorf("test %d; expected error, got none", i)
		} else if err != nil && !test.expErr {
			t.Errorf("test %d; unexpected error; %v", i, err)
		}
		if prePerformTest != test.expPrePerform {
			t.Errorf(
				"test %d; prePerformCheck incorrect; got %q, expected %q",
				i, prePerformTest, test.expPrePerform,
			)
		}
		if usage != test.expUsage {
			t.Errorf("test %d; got %q, expected %q", i, usage, test.expUsage)
			t.Log(test.args)
		}
	}
}

func TestDelegator(t *testing.T) {
	t.Run("DescribeSubcommands", func(t *testing.T) {
		root := newStubRoot(t.Name(), nil)
		out := root.Delegator.DescribeSubcommands()

		expectedMentions := []string{"alpha", "bravo", "charlie", "delta"}
		if len(out) != len(expectedMentions) {
			t.Fatalf(
				"wrong output length; got %d, expected %d",
				len(out), len(expectedMentions),
			)
		}
		for i, expected := range expectedMentions {
			if !strings.Contains(out[i], expected) {
				t.Errorf("item %d; did not mention subcommand name", i)
			}
		}
	})

	t.Run("flags", func(t *testing.T) {
		tests := []struct {
			args              []string
			expectedFlagNames []string
		}{
			{
				args:              []string{"delta", "echo"},
				expectedFlagNames: []string{"bar", "quebec"},
			},
			{
				args:              []string{"delta", "foxtrot"},
				expectedFlagNames: []string{"qux"},
			},
			{
				args:              []string{"delta", "golf"},
				expectedFlagNames: []string{"bar"},
			},
		}
		for i, test := range tests {
			root := newStubRoot(t.Name(), nil)
			delta, ok := root.Subs["delta"].(*alf.Delegator)
			if !ok {
				t.Fatalf(
					"test %d; delta to be a %T",
					i, &alf.Delegator{},
				)
			}
			subDelta, ok := delta.Subs[test.args[1]].(*alf.Command)
			if !ok {
				t.Fatalf(
					"test %d; expected delta subcommand to be a %T",
					i, &alf.Command{},
				)
			}

			childFlagNames := make([]string, 0)
			// Mimics one part of (*Delegator).Perform. Kind of hacky, but no
			// other way to access a Command's flags from here.
			subDelta.Setup(*delta.Flags).VisitAll(
				func(f *flag.Flag) { childFlagNames = append(childFlagNames, f.Name) },
			)

			if len(childFlagNames) != len(test.expectedFlagNames) {
				t.Errorf(
					"test %d; wrong number of child flags; got %d, expected %d",
					i, len(childFlagNames), len(test.expectedFlagNames),
				)
			}
			for j, name := range childFlagNames {
				if name != test.expectedFlagNames[j] {
					t.Errorf(
						"test %d %d; wrong name; got %q, expected %q",
						i, j, name, test.expectedFlagNames[j],
					)
				}
			}
		}
	})
}

func TestCommand(t *testing.T) {
	root := newStubRoot(t.Name(), nil)

	alpha := root.Subs["alpha"].Perform(context.TODO())
	if alpha != nil {
		t.Errorf("should return result of Run; got %v, expected %v", alpha, nil)
	}

	bravo := root.Subs["bravo"].Perform(context.TODO())
	if bravo != errStub {
		t.Errorf("should return result of Run; got %v, expected %v", bravo, errStub)
	}
}
