#!/usr/bin/env -S just -f

GO := "go"
GOSEC := "gosec"
PKG_PATH := "./..."

# list recipes
[default]
@default:
    just -f {{ justfile() }} --list --unsorted

# sanity check for compilation errors
[group('build')]
build:
    {{ GO }} build {{ PKG_PATH }}

# compile example
[group('build')]
build-examples:
    mkdir -pv bin && {{ GO }} build -o ./bin/full_example ./examples/full

# get module dependencies, tidy them up
[group('build')]
mod-tidy:
    {{ GO }} mod tidy

# run tests (override variable value ARGS to use test flags)
[group('test')]
test *args:
    {{ GO }} test {{ PKG_PATH }} {{ args }}

# examine source code for suspicious constructs
[group('static-analysis')]
vet *args:
    {{ GO }} vet {{ args }} {{ PKG_PATH }}

[doc("Run a security scanner over the source code. This justfile won't install
the binary for you, so check out the gosec README for instructions:
    https://github.com/securego/gosec
If necessary, specify the path to the binary with the GOSEC variable.
    $ just GOSEC=path/to/gosec gosec
    $ just --set GOSEC path/to/gosec gosec")]
[group('static-analysis')]
gosec *args:
    {{ GOSEC }} {{ args }} {{ PKG_PATH }}
