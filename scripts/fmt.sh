#!/bin/sh
test -d .git  || echo "run this script from root project directory" && exitcode=1
for file in `git ls-tree -r master --name-only | grep   \\\.go`
do
	output=`gofmt -w "$file"`

	if test -n "$output"
	then
		echo >&2 "$output"
		exitcode=1
	fi
done
