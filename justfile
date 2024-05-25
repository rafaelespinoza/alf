#!/usr/bin/env -S just -f

GO := "go"
GOSEC := "gosec"
PKG_IMPORT_PATH := "github.com/rafaelespinoza/alf"

# list recipes
@default:
    just -f {{ justfile() }} --list --unsorted

# sanity check for compilation errors
build:
	{{ GO }} build {{ PKG_IMPORT_PATH }}

# compile example
build-examples:
	mkdir -pv bin && {{ GO }} build -o ./bin/full_example ./examples/full

# get module dependencies, tidy them up
mod:
    {{ GO }} mod tidy

# run tests (override variable value ARGS to use test flags)
test ARGS='':
    {{ GO }} test {{ PKG_IMPORT_PATH }}/... {{ ARGS }}

# examine source code for suspicious constructs
vet ARGS='':
    {{ GO }} vet {{ ARGS }} {{ PKG_IMPORT_PATH }}/...

# Run a security scanner over the source code. This justfile won't install the
# scanner binary for you, so check out the gosec README for instructions:
# https://github.com/securego/gosec
#
# If necessary, specify the path to the built binary with the GOSEC env var.
#
# Also note, the package paths (last positional input to gosec command) should
# be a "relative" package path. That is, starting with a dot.
gosec ARGS='':
	{{ GOSEC }} {{ ARGS }} ./...
