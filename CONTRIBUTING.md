# Contributing to ShellHub

ShellHub is an open source project and we love to receive
contributions from our community.

Here are a few general guidelines on contributing and reporting bugs
to ShellHub that we ask you to take a look first.  Notice that all of
your interactions in the project are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Reporting Issues

Before reporting a new issue, please be sure that the issue wasn't
already reported or fixed by searching on GitHub through our
[issues](https://github.com/ShellHub-io/shellhub/issues).


## Development environment

Everything runs in containers, so the host needs only Docker with Compose and git. Use the
`bin/docker-compose` wrapper rather than `docker compose` directly. The wrapper picks the compose
files for the edition and environment you set.

Switch the stack to development mode, which adds hot reload and a built-in agent, and start it.
`make start` generates the service keys on the first run:

```sh
echo "SHELLHUB_ENV=development" >> .env.override
make start
```

Once the services are up, create a user and a namespace. Keep the tenant ID as written,
because the built-in agent registers into it.

```sh
./bin/cli user create <username> <password> <email>
./bin/cli namespace create <namespace> <username> 00000000-0000-4000-0000-000000000000
```

Open `http://localhost` and accept the pending device. Open a shell on it from the console, or with
your own SSH key through `ssh shellhub@<namespace>.shellhub@localhost`. The first time, ShellHub
prints a link to confirm the key in the console. The session lands in the agent's container, not
on your machine. [devscripts](devscripts) has helpers for adding more devices and running tests.

## Submitting Changes

* Check for open issues, or open a fresh issue to start a discussion
  around a feature idea or a bug. Opening a separate issue to discuss
  the change is less important for smaller changes, as the discussion
  can be done in the pull request.
* Fork the relevant repository on GitHub, and start making your
  changes.
* Check out the README for the project for specific information to
  that repository.
* Run `devscripts/comment-scan --diff` before pushing. The code style
  allows a comment only as a doc comment on an exported declaration or as
  the reason on a linter suppression; CI runs the same check on every pull
  request.
* Push the change to a separate branch for your feature.
* Open a pull request.
* We try to merge and deploy changes as soon as possible, or at least
  leave some feedback.
