#!/bin/sh

cleanup() {
    exit 0
}

trap cleanup SIGTERM SIGINT

air &

wait $!
