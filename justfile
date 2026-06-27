#!/usr/bin/env -S just -f

GO := "go"
GOSEC := "gosec"
PKG_PATH := "./..."

[private]
_BIN_DIR := 'bin'

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
build-examples: _mk_bin_dir
    {{ GO }} build -o {{ _BIN_DIR }}/alf-example ./examples/full

@_mk_bin_dir:
    mkdir -pv {{ _BIN_DIR }}

# get module dependencies, tidy them up
[group('build')]
mod-tidy:
    {{ GO }} mod tidy

# run tests (override variable value ARGS to use test flags)
[group('test')]
test *args:
    {{ GO }} test {{ PKG_PATH }} {{ args }}

[group('test')]
build-integration-testbin: _mk_bin_dir
    {{ GO }} build -o {{ _BIN_DIR }}/alf-example.test -cover ./examples/full

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

CONTAINER_TOOL := 'podman'
TEST_CONTAINER_NAME := 'localhost/alf_test'

# make container image for integration testing
[group('test')]
[group('container')]
build-test-image:
    {{ CONTAINER_TOOL }} image build -t {{ TEST_CONTAINER_NAME }} -f integration_tests/Containerfile .

# run integration tests in a container
[group('test')]
[group('container')]
run-test-container *run_flags:
    {{ CONTAINER_TOOL }} run {{ run_flags }} {{ TEST_CONTAINER_NAME }}
