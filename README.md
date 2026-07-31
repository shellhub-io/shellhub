<h1 align="center">
  <a href="https://shellhub.io"><img src="http://docs.shellhub.io/img/logo.png" alt="ShellHub"></a>
</h1>

<h4 align="center">Centralized SSH for the edge and cloud computing.</h4>

<p align="center">
  <a href="https://github.com/shellhub-io/shellhub/actions?query=workflow%3AQA">
    <img src="https://github.com/shellhub-io/shellhub/workflows/QA/badge.svg" alt="GitHub Workflow">
  </a>
  <a href="https://gitter.im/shellhub-io/community">
    <img src="https://badges.gitter.im/shellhub-io/shellhub.svg">
  </a>
</p>

<p align="center">
  <a href="https://cloud.shellhub.io">ShellHub Cloud</a> •
  <a href="http://docs.shellhub.io">Documentation</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#getting-help">Getting Help</a>
</p>

ShellHub is a centralized SSH gateway that allows users to remotely access and
manage their servers and devices from anywhere, using a web browser or mobile app.
It offers a secure and convenient way to connect and control your servers and devices.

One of the main benefits of ShellHub is that it acts as a central gateway for all your
Linux servers and devices, allowing you to access them from anywhere with an internet connection.
This means you don't have to worry about getting its public IP address, configuring the router,
changing VPN/firewall settings or using a jump host to access your servers and devices.
This can be inconvenient and time-consuming.

ShellHub also allows you to access and manage multiple servers and devices from a single interface,
saving time and making it easier to keep track of all your servers and devices.
The platform also includes enhanced security features such as public key authentication,
SSH firewall rules to prevent unauthorized access, audit logging, and session recording
to provide a record of activity for compliance purposes.

Another benefit of ShellHub is its ability to work with a variety of devices,
including embedded Linux devices. This makes it a versatile tool for a wide range of applications,
from managing servers in a data center to controlling industrial equipment or Internet of Things (IoT) devices.

Overall, ShellHub is a powerful tool for managing and securing your servers and devices,
whether you are a small business, a large enterprise, or an individual user.

![Screenshot](https://github.com/shellhub-io/shellhub-io.github.io/raw/src/docs/img/screenshot.png)

## Features

### :computer: Native SSH support

ShellHub provides native SSH support, allowing you to access any device behind the ShellHub SSH gateway using standard tools such as OpenSSH Client and PuTTY. This means you don't need to install any additional third-party tools to connect to your devices behind ShellHub. Simply use the tools you are already familiar with to remotely manage your servers and devices through ShellHub.

### :file_folder: SCP/SFTP support

ShellHub offers SCP/SFTP support, allowing you to copy files to and from your devices using industry standard tools without the need for any additional third-party applications. This makes it easy to transfer files securely to and from your servers and devices, ensuring that your data remains safe and secure throughout the process.

### :arrows_counterclockwise: SSH port forwarding

ShellHub allows you to securely forward TCP traffic from a local port on your machine to a remote port on a device behind the ShellHub gateway. This can be useful for a variety of purposes, such as accessing remote services that are not directly accessible from your local machine, securely transferring data between two hosts over an untrusted network and use SOCKS Proxy. 

### :key: Public-key authentication

ShellHub supports public-key authentication, which allows multiple users to log in as the same system user without having to share a single password. This can make it easier to manage access for multiple users, as you can revoke a single user's access without affecting the access of other users. Additionally, public-key authentication can make it easier for a single user to log in to many accounts without having to manage multiple passwords.

### :shield: Firewall rules

ShellHub provides flexible firewall rules for filtering SSH connections, giving you fine-grained control over which SSH connections can reach your devices. This helps to improve the security of your servers and devices by allowing you to specify which connections are allowed and which are blocked, helping to prevent unauthorized access.

### :spiral_notepad: Audit logging

ShellHub includes audit logging capabilities, which means that every time an SSH connection is made to ShellHub, a session is created and stored on the server for audit purposes. This can be useful for tracking and monitoring access to your servers and devices, and can help you to identify unauthorized access attempts.

### :movie_camera: Session recording

ShellHub offers session recording, which means that all interactive SSH sessions are recorded, including all user activity that occurs during the session. These recordings can then be replayed via a built-in session player in the ShellHub Web UI. This feature can be useful for a variety of purposes, such as training and documentation, as well as for tracking and monitoring user activity on your servers and devices.

### :whale: Container Remote Access

ShellHub seamlessly integrates with Docker, enabling you to remotely access Docker containers.
With this feature, users can securely connect to containers.
Whether you're troubleshooting, performing maintenance, or overseeing your containerized services,
ShellHub's Container Remote Access simplifies container remote access management,
providing flexibility and control over containerized environments.

## Getting started

To self-host ShellHub on your own servers and managing your own infrastructure
(installation, maintenance, upgrades, backup and so on) follow our
[Self-Hosting Guide](https://docs.shellhub.io/self-hosted/deploying).

If you prefer to use the cloud hosted service where we manage everything for your ease
and convenience, create a free account in [ShellHub Cloud](https://cloud.shellhub.io).

## Contributing

ShellHub is an open source project and we love to receive contributions from
our community. If you would like to contribute, please read
our [contributing guide](CONTRIBUTING.md).

## Getting Help

You are welcome to submit any questions, participate in discussions and request
help with any issue in our [Gitter Channel](https://gitter.im/shellhub-io/community).

## License

ShellHub is licensed under the Apache License, Version 2.0.
See [LICENSE](LICENSE.md) for the full license text.

## Development Environment Setup

First of all, you'll need to generate required keys for the services:

```
$ make keygen
```

Set the environment to development:

```
$ echo "SHELLHUB_ENV=development" >> .env.override
```

Start the whole environment:

```
$ make start
```

> Avoid to use `docker-compose` directly (unless you know what you're doing); instead use `bin/docker-compose` wrapper.

Wait for all services to be ready then create initial user by running:

```
$ ./bin/cli user create <username> <password> <email>
```

```
$ ./bin/cli namespace create <namespace> <owner> 00000000-0000-4000-0000-000000000000
```

> Don't change the value of `00000000-0000-4000-0000-000000000000` (the TenantID), this value is hardcoded during agent initialization in development mode.

When you open ShellHub UI for the first time, be sure to accept pending device.

See the [devscripts which can be useful for development](./devscripts).

## Authors

ShellHub was created by [O.S. Systems](https://www.ossystems.com.br).

## Contributors

Thanks to everyone who has contributed to ShellHub. The full list lives in the
[contributor graph](https://github.com/shellhub-io/shellhub/graphs/contributors).
