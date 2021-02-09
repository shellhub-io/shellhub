# Devscripts

This directory contains scripts which can be useful for development.
They are not intented for regular use by end users.

## Requirements

* Bash (although we try to avoid bashism in devscripts, things can happen)
* docker-compose
* HTTPie (https://httpie.org)
* jq (https://stedolan.github.io/jq)

## Scripts

* `add-device`: Add a random fake device to ShellHub
* `get-devices`: Get devices from API
* `lint-code`: Run code linter
* `test-unit`: Run unit test
* `gen-mock`: Generate/update mock objects for testing
