#!/bin/bash

set -euo pipefail

_base_dir="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
source $_base_dir/vendor/github.com/reconquest/import.bash/import.bash

:import:source github.com/reconquest/test-runner.bash
:import:source github.com/reconquest/blank.bash

test-runner:set-local-setup      util/setup.sh
test-runner:set-local-teardown   util/teardown.sh
test-runner:set-testcases-dir    testcases

test-runner:run "${@}"
