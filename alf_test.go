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

func newStubRoot(name string) alf.Root {
	var (
		foo string
		bar int
		qux bool
	)

	charlie := alf.Delegator{
		Description: "subcmd with more subs",
		Flags:       newMutedFlagSet("charlie", flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"delta": &alf.Command{
				Description: "decorate (mutate) the input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					p.Init("charlie delta", flag.ContinueOnError)
					p.BoolVar(&qux, "quebec", true, "qqq")
					return &p
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"echo": &alf.Command{
				Description: "ignore input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					flags := newMutedFlagSet("charlie echo", flag.ContinueOnError)
					flags.BoolVar(&qux, "qux", false, "QQQ")
					return flags
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"foxtrot": &alf.Command{
				Description: "allow input flags to pass through",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					return &p
				},
				Run: func(ctx context.Context) error { return nil },
			},
		},
	}
	charlie.Flags.IntVar(&bar, "bar", 2, "bbb")

	root := alf.Delegator{
		Description: "root",
		Flags:       newMutedFlagSet(name, flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"alpha": &alf.Command{
				Description: "ok command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					return newMutedFlagSet("alpha", flag.ContinueOnError)
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"bravo": &alf.Command{
				Description: "error command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					return newMutedFlagSet("bravo", flag.ContinueOnError)
				},
				Run: func(ctx context.Context) error { return errStub },
			},
			"charlie": &charlie,
		},
	}
	root.Flags.StringVar(&foo, "foo", "foxtrot", "fff")

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
	tests := []struct {
		args   []string
		expErr bool
	}{
		{args: []string{""}, expErr: true},
		{args: []string{"-h"}, expErr: true},
		{args: []string{"--help"}, expErr: true},
		{args: []string{"alpha"}, expErr: false},
		{args: []string{"bravo"}, expErr: true},
		{args: []string{"charlie"}, expErr: true},
		{args: []string{"charlie", "-h"}, expErr: true},
		{args: []string{"charlie", "delta"}, expErr: false},
		{args: []string{"charlie", "echo"}, expErr: false},
		{args: []string{"charlie", "foxtrot"}, expErr: false},
		{args: []string{"charlie", "golf"}, expErr: true},
		{args: []string{"charlie", "delta", "-h"}, expErr: true},
		{args: []string{"charlie", "echo", "-h"}, expErr: true},
		{args: []string{"charlie", "foxtrot", "-h"}, expErr: true},
	}

	for i, test := range tests {
		root := newStubRoot(fmt.Sprintf("%s %d", t.Name(), i))

		err := root.Run(context.TODO(), test.args)
		if err == nil && test.expErr {
			t.Errorf("test %d; expected error, got none", i)
		} else if err != nil && !test.expErr {
			t.Errorf("test %d; unexpected error; %v", i, err)
		}
	}
}

func TestDelegator(t *testing.T) {
	t.Run("DescribeSubcommands", func(t *testing.T) {
		root := newStubRoot(t.Name())
		out := root.Delegator.DescribeSubcommands()

		expectedMentions := []string{"alpha", "bravo", "charlie"}
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
				args:              []string{"charlie", "delta"},
				expectedFlagNames: []string{"bar", "quebec"},
			},
			{
				args:              []string{"charlie", "echo"},
				expectedFlagNames: []string{"qux"},
			},
			{
				args:              []string{"charlie", "foxtrot"},
				expectedFlagNames: []string{"bar"},
			},
		}
		for i, test := range tests {
			root := newStubRoot(t.Name())
			charlie, ok := root.Subs["charlie"].(*alf.Delegator)
			if !ok {
				t.Fatalf(
					"test %d; charlie to be a %T",
					i, &alf.Delegator{},
				)
			}
			subCharlie, ok := charlie.Subs[test.args[1]].(*alf.Command)
			if !ok {
				t.Fatalf(
					"test %d; expected charlie subcommand to be a %T",
					i, &alf.Command{},
				)
			}

			childFlagNames := make([]string, 0)
			// Mimics one part of (*Delegator).Perform. Kind of hacky, but no
			// other way to access a Command's flags.
			subCharlie.Setup(*charlie.Flags).VisitAll(
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
	root := newStubRoot(t.Name())

	alpha := root.Subs["alpha"].Perform(context.TODO())
	if alpha != nil {
		t.Errorf("should return result of Run; got %v, expected %v", alpha, nil)
	}

	bravo := root.Subs["bravo"].Perform(context.TODO())
	if bravo != errStub {
		t.Errorf("should return result of Run; got %v, expected %v", bravo, errStub)
	}
}
