#!/bin/sh

cleanup() {
    exit 0
}

# Trap SIGTERM and SIGINT to ensure cleanup
trap cleanup SIGTERM SIGINT

# Start air in background
air &

# Wait for air process
wait $!
