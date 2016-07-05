#!/bin/bash

set -euo pipefail

_base_dir="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
source $_base_dir/vendor/github.com/reconquest/import.bash/import.bash

:import:source github.com/reconquest/test-runner.bash
:import:source github.com/reconquest/blank.bash

test-runner:set-local-setup      tests/util/setup.bash
test-runner:set-local-teardown   tests/util/teardown.bash
test-runner:set-testcases-dir    tests/testcases

test-runner:run "${@}"
