#!/bin/sh

print_usage() {
    echo "Usage: $0 <username> <password> <server> <operation>"
    echo "       operation: exec"
}

if [ "$#" -ne 4 ]; then
    echo "Error: Invalid number of arguments"
    print_usage
    exit 1
fi

user=$1
password=$2
server=$3
operation=$4

case $operation in
    "exec")
        counter=0

        echo "Running SSH stress test using \"exec\" method"

        while true; do
            echo "Iteration number $counter"

            output=$(sshpass -p "$password" ssh "$user@$server" uname)
            if [ -z "$output" ]; then
                echo "Test command run returned no content"

                exit 1
            fi

            counter=$((counter+1))
        done
        ;;
    *)
        echo "Error: Invalid operation"
        print_usage
        exit 1
        ;;
esac

