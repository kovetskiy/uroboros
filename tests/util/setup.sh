#!/bin/bash

alias @var='tests:value'

tests:clone vendor vendor
tests:clone util/stash bin/stash

touch stash

:stash-start() {
    blank:server:new stash
    tests:eval $stash::set-program stash
    tests:eval $stash::start
    tests:eval $stash::get-pid

    tests:put-string stash "$($stash::get-dir)"
}

:stash-stop() {
    $stash:stop
}

:uroboros-port() {
    if [ -f port ]; then
        cat port
        return
    fi

    local number=$((10000+$RANDOM))

    tests:put-string port "$number"

    echo "$number"
}

:uroboros-configure() {

}

:uroboros-start() {
    @var port :uroboros-port

    tests:run-background task $BUILD -c config ${@}
    tests:put-string "task" "$task"

    @var uroboros_process tests:get-background-pid $task

    local i=0
    while :; do
        tests:describe "waiting for uroboros listening task"
        sleep 0.05

        local netstat=""
        if netstat=$(netstat -na | grep ":$port"); then
            tests:describe "$netstat"
            break
        fi

        i=$((i+1))
        if [ "$i" -gt 20 ]; then
            tests:fail "process doesn't started listening at $port"
        fi
    done
}
