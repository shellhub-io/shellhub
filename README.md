# ShellHub [![ShellHub](https://circleci.com/gh/shellhub-io/shellhub.svg?style=shield)](https://circleci.com/gh/shellhub-io/shellhub)

ShellHub is a modern SSH server for remotely accessing Linux devices
via command line (using any SSH client) or web-based user interface, 
designed as an alternative to _sshd_. It enables teams to
easily access any Linux device behind firewall and NAT.

![Screenshot](https://github.com/shellhub-io/shellhub-io.github.io/raw/src/docs/img/screenshot.png)

## Getting started

To start using ShellHub, it is suggested that you follow the
[Getting started](https://shellhub-io.github.io/getting-started/) guide 
in the ShellHub documentation.

## Contributing

ShellHub is an open source project and we love to receive contributions from
our community. If you would like to contribute, please read
our [contributing guide](CONTRIBUTING.md).

## Getting Help

You are welcome to submit any questions, participate in discussions and request
help with any issue in our [Gitter Channel](https://gitter.im/shellhub-io/community).

## License

ShellHub is licensed under the Apache License, Version 2.0.
See [LICENSE](LICENSE) for the full license text.

## Development Environment Setup

First of all, you'll need to generate required keys for the services:

```
$ ./bin/keygen
```

Start the whole environment:

```
$ docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

Wait for all services to be ready then create initial user running:

```
$ ./bin/add-user <username> <password>
```

## Authors

ShellHub was created by [O.S. Systems](https://www.ossystems.com.br).
