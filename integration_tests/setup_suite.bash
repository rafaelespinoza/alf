#!/usr/bin/env bats
 
set -eu -o pipefail

# setup_suite is a lifecycle function, picked up by bats.
# The file must be named setup_suite.bash.
# Bash testing library, bats. https://bats-core.readthedocs.io/.
function setup_suite() {
  # Reference reads for go integration test coverage.
  # https://go.dev/blog/integration-test-coverage
  # https://go.dev/doc/build-cover
  : "${GOCOVERDIR:?GOCOVERDIR is required}"
  if [[ ! -d "${GOCOVERDIR}" ]]; then
  	mkdir -pv "${GOCOVERDIR}"
  fi

  bats_load_library bats-assert
  bats_load_library bats-file
  bats_load_library bats-support

  assert_file_executable /usr/local/bin/alf-example.test
}

# teardown_suite is another lifecycle function for bats.
# It processes test coverage data.
function teardown_suite() {
  : "${GOCOVERDIR:?GOCOVERDIR is required}"

  local -r combined_coverage_file="${GOCOVERDIR}/integration_test_coverage.out"
  mkdir -pv "${GOCOVERDIR}/merged"
  go tool covdata merge -i "${GOCOVERDIR}" -o "${GOCOVERDIR}/merged"
  go tool covdata textfmt -i "${GOCOVERDIR}/merged" -o "${combined_coverage_file}"
  echo >&3 "# wrote coverage to ${combined_coverage_file}"
}
