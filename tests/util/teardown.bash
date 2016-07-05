#!/bin/bash

blank:server:new stash $(cat stash)

if [[ "$($stash::get-pid)" ]]; then
    tests:describe stash logs
    tests:eval $stash::logs
    tests:silence tests:eval $stash::cleanup
fi
