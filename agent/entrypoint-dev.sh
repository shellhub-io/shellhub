#!/bin/sh

# Cleanup function to kill Delve processes on exit
cleanup() {
    echo "Cleaning up Delve processes..."
    pkill -9 dlv
    exit 0
}

# Trap SIGTERM and SIGINT to ensure cleanup
trap cleanup SIGTERM SIGINT

# Start air in background
air &

# Wait for air process
wait $!
