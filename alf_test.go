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

func newStubRoot(name string, usage *string, timesUsageCalled *int) alf.Root {
	var (
		foo string
		bar int
		qux bool
	)

	// onUsage captures invocations of a Usage function. It helps with checking
	// the correct Usage func is called, and that any Usage function is called
	// just once.
	onUsage := func(val string) {
		if usage == nil || *usage == "" {
			*usage = val
		}
		if timesUsageCalled != nil {
			*timesUsageCalled++
		}
	}

	india := alf.Delegator{
		Description: "subcmd of a subcmd with more subs",
		Flags:       newMutedFlagSet("india", flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"foo": &alf.Command{
				Description: "works OK",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("delta india foo", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.delta.india.foo") }
					return f
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"bar": &alf.Command{
				Description: "has an error",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("delta india bar", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.delta.india.bar") }
					return f
				},
				Run: func(ctx context.Context) error { return errStub },
			},
		},
	}
	india.Flags.Usage = func() { onUsage("root.delta.india") }

	delta := alf.Delegator{
		Description: "subcmd with more subs",
		Flags:       newMutedFlagSet("delta", flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"echo": &alf.Command{
				Description: "decorate (mutate) the input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					p.Init("delta echo", flag.ContinueOnError)
					p.BoolVar(&qux, "quebec", true, "qqq")
					p.Usage = func() { onUsage("root.delta.echo") }
					return &p
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"foxtrot": &alf.Command{
				Description: "ignore input flags",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("delta foxtrot", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.delta.foxtrot") }
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
					f.Usage = func() { onUsage("root.delta.hotel") }
					return f
				},
				Run: func(ctx context.Context) error { return alf.ErrShowUsage },
			},
			"india": &india,
		},
	}
	delta.Flags.IntVar(&bar, "bar", 2, "bbb")
	delta.Flags.Usage = func() { onUsage("root.delta") }

	root := alf.Delegator{
		Description: "root",
		Flags:       newMutedFlagSet(name, flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"alpha": &alf.Command{
				Description: "ok command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("alpha", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.alpha") }
					return f
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"bravo": &alf.Command{
				Description: "error command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("bravo", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.bravo") }
					return f
				},
				Run: func(ctx context.Context) error { return errStub },
			},
			"charlie": &alf.Command{
				Description: "force show help command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("charlie", flag.ContinueOnError)
					f.Usage = func() { onUsage("root.charlie") }
					return f
				},
				Run: func(ctx context.Context) error { return alf.ErrShowUsage },
			},
			"delta": &delta,
		},
	}
	root.Flags.StringVar(&foo, "foo", "frank", "root flag example")
	root.Flags.Usage = func() { onUsage("root") }

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
	type testCase struct {
		args          []string
		prePerform    func(context.Context) error
		expPrePerform string
		expErr        bool
		expUsage      string
	}

	runTest := func(t *testing.T, test testCase) {
		t.Helper()

		prePerformTest = ""
		var usage string
		var usageCalls int
		root := newStubRoot(t.Name(), &usage, &usageCalls)
		if test.prePerform != nil {
			root.PrePerform = test.prePerform
		}

		err := root.Run(context.TODO(), test.args)
		if err == nil && test.expErr {
			t.Error("expected error, got none")
		} else if err != nil && !test.expErr {
			t.Errorf("unexpected error; %v", err)
		}

		if prePerformTest != test.expPrePerform {
			t.Errorf(
				"prePerformCheck incorrect; got %q, expected %q",
				prePerformTest, test.expPrePerform,
			)
		}

		if usage != test.expUsage {
			t.Errorf(
				"wrong Usage func; got %q, expected %q; input args: %v",
				usage, test.expUsage, test.args,
			)
		}

		if usageCalls > 1 {
			t.Errorf("called Usage %d times; expected <= 1", usageCalls)
		}
	}

	t.Run("Returns whatever the Command returns", func(t *testing.T) {
		runTest(t, testCase{args: []string{"alpha"}, expErr: false})
		runTest(t, testCase{args: []string{"bravo"}, expErr: true})
	})

	t.Run("Calls correct Usage function", func(t *testing.T) {
		// Root
		runTest(t, testCase{args: []string{""}, expErr: true, expUsage: "root"})
		runTest(t, testCase{args: []string{"-h"}, expErr: true, expUsage: "root"})
		runTest(t, testCase{args: []string{"--help"}, expErr: true, expUsage: "root"})

		// Command
		runTest(t, testCase{args: []string{"charlie"}, expErr: true, expUsage: "root.charlie"})

		// Delegator
		runTest(t, testCase{args: []string{"delta"}, expErr: true, expUsage: "root.delta"})
		runTest(t, testCase{args: []string{"delta", "-h"}, expErr: true, expUsage: "root.delta"})

		// Delegator -> Command
		runTest(t, testCase{args: []string{"delta", "echo", "-h"}, expErr: true, expUsage: "root.delta.echo"})
		runTest(t, testCase{args: []string{"delta", "foxtrot", "-h"}, expErr: true, expUsage: "root.delta.foxtrot"})
		runTest(t, testCase{args: []string{"delta", "golf", "-h"}, expErr: true, expUsage: "root.delta"})
		runTest(t, testCase{args: []string{"delta", "hotel"}, expErr: true, expUsage: "root.delta.hotel"})

		// Delegator -> Delegator
		runTest(t, testCase{args: []string{"delta", "india"}, expErr: true, expUsage: "root.delta.india"})

		// Delegator -> Delegator -> Command
		runTest(t, testCase{args: []string{"delta", "india", "foo", "-h"}, expErr: true, expUsage: "root.delta.india.foo"})
		runTest(t, testCase{args: []string{"delta", "india", "bar", "-h"}, expErr: true, expUsage: "root.delta.india.bar"})
	})

	t.Run("Parent flags", func(t *testing.T) {
		// These test really just document where parent command flags are
		// supposed to be specified.

		// A root command's flag should be before the subcommand name.
		runTest(t, testCase{args: []string{"-foo", "fred", "alpha"}, expErr: false})

		// But if the root command flag is specified after the subcommand name,
		// then it's interpreted as a subcommand flag. This example results in
		// an error because the subcommand doesn't define this flag.
		runTest(t, testCase{args: []string{"alpha", "-foo", "fff"}, expErr: true, expUsage: "root.alpha"})
	})

	t.Run("PrePerform", func(t *testing.T) {
		runTest(t, testCase{
			args: []string{"alpha"},
			prePerform: func(ctx context.Context) error {
				prePerformTest = "alpha prePerformCheck"
				return nil
			},
			expErr:        false,
			expPrePerform: "alpha prePerformCheck",
		})

		// When an error occurs in PrePerform
		runTest(t, testCase{
			args:       []string{"delta", "foxtrot"},
			prePerform: func(ctx context.Context) error { return errStub },
			expErr:     true,
		})

		runTest(t, testCase{
			args: []string{"delta", "foxtrot"},
			prePerform: func(ctx context.Context) error {
				return fmt.Errorf("%w, wrap", alf.ErrShowUsage)
			},
			expErr:   true,
			expUsage: "root",
		})
	})

	t.Run("Unknown command", func(t *testing.T) {
		runTest(t, testCase{args: []string{"echo"}, expErr: true, expUsage: "root"})
		runTest(t, testCase{args: []string{"delta", "zulu"}, expErr: true, expUsage: "root.delta"})
	})
}

func TestDelegator(t *testing.T) {
	t.Run("DescribeSubcommands", func(t *testing.T) {
		root := alf.Delegator{
			Description: "root",
			Flags:       newMutedFlagSet("root", flag.ContinueOnError),
			Subs: map[string]alf.Directive{
				"alpha": &alf.Command{
					Description: "a",
					Setup:       func(p flag.FlagSet) *flag.FlagSet { return &p },
					Run:         func(ctx context.Context) error { return nil },
				},
				"bravo": &alf.Command{
					Description: "b",
					Setup:       func(p flag.FlagSet) *flag.FlagSet { return &p },
					Run:         func(ctx context.Context) error { return nil },
				},
				"charlie": &alf.Command{
					Description: "c",
					Setup:       func(p flag.FlagSet) *flag.FlagSet { return &p },
					Run:         func(ctx context.Context) error { return nil },
				},
				"delta": &alf.Command{
					Description: "d",
					Setup:       func(p flag.FlagSet) *flag.FlagSet { return &p },
					Run:         func(ctx context.Context) error { return nil },
				},
			},
		}
		out := root.DescribeSubcommands()

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

	// Tests that a Delegator with Subs can pass flags from a parent to child in
	// various ways.
	t.Run("sharing flag data", func(t *testing.T) {
		type testCase struct {
			delegator         alf.Delegator
			subcmdName        string
			expectedFlagNames []string
		}

		runTest := func(t *testing.T, test testCase) {
			t.Helper()

			subcmd := test.delegator.Subs[test.subcmdName].(*alf.Command)
			childFlagNames := make([]string, 0)
			// Mimics one part of (*Delegator).Perform. Kind of hacky, but no
			// other way to access a Command's flags from here.
			subcmd.Setup(*test.delegator.Flags).VisitAll(
				func(f *flag.Flag) { childFlagNames = append(childFlagNames, f.Name) },
			)

			if len(childFlagNames) != len(test.expectedFlagNames) {
				t.Errorf(
					"wrong number of child flags; got %d, expected %d",
					len(childFlagNames), len(test.expectedFlagNames),
				)
			}
			for i, name := range childFlagNames {
				if name != test.expectedFlagNames[i] {
					t.Errorf(
						"flag[%d]; wrong name; got %q, expected %q",
						i, name, test.expectedFlagNames[i],
					)
				}
			}
		}

		makeTestDelegator := func() alf.Delegator {
			root := alf.Delegator{
				Description: "root",
				Flags:       newMutedFlagSet("root", flag.ContinueOnError),
			}
			var (
				bar int
				qux bool
			)
			root.Flags.IntVar(&bar, "bar", 2, "bbb")

			root.Subs = map[string]alf.Directive{
				"alpha": &alf.Command{
					Description: "decorate (mutate) the input flags",
					Setup: func(p flag.FlagSet) *flag.FlagSet {
						p.Init("a", flag.ContinueOnError)
						p.BoolVar(&qux, "quebec", true, "qqq")
						return &p
					},
					Run: func(ctx context.Context) error { return nil },
				},
				"bravo": &alf.Command{
					Description: "ignore input flags",
					Setup: func(p flag.FlagSet) *flag.FlagSet {
						f := newMutedFlagSet("a", flag.ContinueOnError)
						f.BoolVar(&qux, "qux", false, "QQQ")
						return f
					},
					Run: func(ctx context.Context) error { return nil },
				},
				"charlie": &alf.Command{
					Description: "allow input flags to pass through",
					Setup: func(p flag.FlagSet) *flag.FlagSet {
						return &p
					},
					Run: func(ctx context.Context) error { return nil },
				},
			}
			return root
		}

		t.Run("decorate parent flags", func(t *testing.T) {
			// uses the parent flags and adds its own flags.
			runTest(t, testCase{
				delegator:         makeTestDelegator(),
				subcmdName:        "alpha",
				expectedFlagNames: []string{"bar", "quebec"},
			})
		})

		t.Run("ignore parent flags", func(t *testing.T) {
			// only considers its own flags.
			runTest(t, testCase{
				delegator:         makeTestDelegator(),
				subcmdName:        "bravo",
				expectedFlagNames: []string{"qux"},
			})
		})

		t.Run("pass-through parent flags", func(t *testing.T) {
			// doesn't need to add its own flags, can just use parent flags.
			runTest(t, testCase{
				delegator:         makeTestDelegator(),
				subcmdName:        "charlie",
				expectedFlagNames: []string{"bar"},
			})
		})
	})
}

func TestCommand(t *testing.T) {
	root := alf.Delegator{
		Description: "root",
		Flags:       newMutedFlagSet("root", flag.ContinueOnError),
		Subs: map[string]alf.Directive{
			"alpha": &alf.Command{
				Description: "ok command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("alpha", flag.ContinueOnError)
					return f
				},
				Run: func(ctx context.Context) error { return nil },
			},
			"bravo": &alf.Command{
				Description: "error command",
				Setup: func(p flag.FlagSet) *flag.FlagSet {
					f := newMutedFlagSet("bravo", flag.ContinueOnError)
					return f
				},
				Run: func(ctx context.Context) error { return errStub },
			},
		},
	}

	got := root.Subs["alpha"].Perform(context.TODO())
	if got != nil {
		t.Errorf("should return result of Run; got %v, expected %v", got, nil)
	}

	got = root.Subs["bravo"].Perform(context.TODO())
	if got != errStub {
		t.Errorf("should return result of Run; got %v, expected %v", got, errStub)
	}
}
