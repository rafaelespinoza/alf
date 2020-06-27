```
          __ ____
  ____ _ / // __/
 / __ `// // /_
/ /_/ // // __/
\__,_//_//_/
```

`alf` is a simple toolset for making command line applications in golang. Some
design goals include:

- helping you organize a command line application so its maintainable.
- minimize dependencies.
- written examples, tests to demonstrate how to use it.

This library is nothing but a thin wrapper around the the very awesome and
stable standard library package, [`flag`](https://golang.org/pkg/flag/).  See
the tests for an idea of how to use this library. Check out the examples for
comments and a quick exploratory demo.

```
go run ./examples/full
```

## overview

There's two basic types, `Delegator` and `Command`. A `Delegator` is a parent of
a set of commands. A `Command` does the work.  Both are abstracted as
`Directive`.

Your program should have a `Root`, which is just a `Delegator` with a `Run`
method. From your package main, commence `Run` like so:

```golang
err = root.Run(context.Background(), os.Args[1:])
```

## limitations

Currently it does not permit sharing flag values from a `Root` to a direct child
command. However, you could share flag values between a `Delegator` (one that's
not a `Root` command) and child commands.
