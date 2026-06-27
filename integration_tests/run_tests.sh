#!/usr/bin/env bats

# This script assumes that the CWD is the project source root directory.
 
set -eu -o pipefail

# Bash testing library, bats. https://bats-core.readthedocs.io/.
bats_load_library bats-assert
bats_load_library bats-file
bats_load_library bats-support

# Check the main command
@test 'main outputs usage message without args' {
  run alf-example.test
  assert_success
  assert_output --partial 'Usage:'
}

@test 'main outputs usage message with args: -h' {
  run alf-example.test -h
  assert_success
  assert_output --partial 'Usage:'
}

@test 'main outputs usage message with unknown command' {
  run alf-example.test cmd_does_not_exist
  assert_output --partial 'unknown command'
}

@test 'main handles flags' {
  run alf-example.test -pre
  assert_output --partial 'called Root.PrePerform'
}

# Check subcommand: foo
@test 'foo outputs usage message with args: -h' {
  run alf-example.test foo -h
  assert_success
  assert_output --partial 'Usage:'
}

@test 'foo outputs test' {
  run alf-example.test foo
  assert_output 'test
test
test
test
test'
}

@test 'foo outputs test 2x' {
  run alf-example.test foo -delta 2
  assert_output 'test
test'
}

@test 'foo outputs test123' {
  run alf-example.test foo -echo test123
  assert_output 'test123
test123
test123
test123
test123'
}

@test 'foo outputs test123 2x' {
  run alf-example.test foo -delta 2 -echo test123
  assert_output 'test123
test123'
}

@test 'foo validation error' {
  run alf-example.test foo -delta 123  
  assert_failure
  assert_output --partial 'delta 123 must be <= 42'
}

@test 'foo outputs can use a flag from main' {
  run alf-example.test -pre foo
  assert_output --partial 'called Root.PrePerform'
  assert_output --partial 'test
test
test
test
test'
}

# Check subcommand: bar
@test 'bar outputs usage message without args' {
  run alf-example.test bar
  assert_success
  assert_output --partial 'Usage:'
}

@test 'bar outputs usage message with args: -h' {
  run alf-example.test bar -h
  assert_success
  assert_output --partial 'Usage:'
}

@test 'bar with unknown command' {
  run alf-example.test bar does_not_exist
  assert_failure
  assert_output --partial 'unknown command'
}

## Check subcommand of subcommand: bar cities
@test 'bar cities outputs usage message' {
  run alf-example.test bar cities -h
  assert_success
  assert_output --partial 'Usage:'
}

@test 'bar cities outputs message' {
  run alf-example.test bar cities
  assert_output 'city: "Accra, Ghana", custom charlie: "parker"'
}

@test 'bar cities with bravo flag' {
  run alf-example.test bar cities -bravo
  assert_output 'city: "Beni Mellal, Morocco", custom charlie: "parker"'
}

@test 'bar cities with charlie flag' {
  run alf-example.test bar cities -charlie 'Chuck Berry'
  assert_output 'city: "Accra, Ghana", custom charlie: "Chuck Berry"'
}

@test 'bar cities handles root flag' {
  run alf-example.test -pre bar cities 
  assert_output --partial 'called Root.PrePerform'
  assert_output --partial 'city: "Accra, Ghana", custom charlie: "parker"'
}

## Check subcommand of subcommand: bar oof
@test 'bar oof returns an error if bravo is true'  {
  run alf-example.test bar oof -bravo
  assert_failure
  assert_output --partial 'demo force show usage'
}

@test 'bar oof with alpha flag' {
  run alf-example.test bar oof -alpha 24
  assert_output --partial '24 years old'
}

@test 'bar oof handles root flag' {
  run alf-example.test -pre bar oof
  assert_output --partial 'called Root.PrePerform'
  assert_output --partial 'your alternative charlie "berry" is 42 years old'
}

## Check subcommand of subcommand with subcommands: bar nested
@test 'bar nested demos nested Delegator' {
  run alf-example.test bar nested -h
  assert_output --partial 'Demo of nested Delegator'
}

@test 'bar nested demos nested Delegator help' {
  run alf-example.test bar nested help
  assert_output --partial 'Demo of nested Delegator'
}

@test 'bar nested alfa' {
  run alf-example.test bar nested alfa
  assert_output --partial 'called bar.nested.alfa'
}

@test 'bar nested bravo' {
  run alf-example.test bar nested bravo
  assert_output --partial 'called bar.nested.bravo'
}

@test 'bar nested handles flags from main command' {
  run alf-example.test -pre bar nested bravo
  assert_output --partial 'called Root.PrePerform'
  assert_output --partial 'called bar.nested.bravo'
}
