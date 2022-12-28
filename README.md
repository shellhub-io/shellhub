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
  <!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="#contributors-"><img src="https://img.shields.io/badge/all_contributors-19-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

<p align="center">
  <a href="https://shellhub.io">ShellHub Cloud</a> •
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

### :key: Public-key authentication

ShellHub supports public-key authentication, which allows multiple users to log in as the same system user without having to share a single password. This can make it easier to manage access for multiple users, as you can revoke a single user's access without affecting the access of other users. Additionally, public-key authentication can make it easier for a single user to log in to many accounts without having to manage multiple passwords.

### :shield: Firewall rules

ShellHub provides flexible firewall rules for filtering SSH connections, giving you fine-grained control over which SSH connections can reach your devices. This helps to improve the security of your servers and devices by allowing you to specify which connections are allowed and which are blocked, helping to prevent unauthorized access.

### :spiral_notepad: Audit logging

ShellHub includes audit logging capabilities, which means that every time an SSH connection is made to ShellHub, a session is created and stored on the server for audit purposes. This can be useful for tracking and monitoring access to your servers and devices, and can help you to identify unauthorized access attempts.

### :movie_camera: Session recording

ShellHub offers session recording, which means that all interactive SSH sessions are recorded, including all user activity that occurs during the session. These recordings can then be replayed via a built-in session player in the ShellHub Web UI. This feature can be useful for a variety of purposes, such as training and documentation, as well as for tracking and monitoring user activity on your servers and devices.

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
$ ./bin/add-user <username> <password> <email>
```

Create a namespace for grouping devices and set user as owner:

```
$ TENANT_ID=00000000-0000-4000-0000-000000000000 ./bin/add-namespace <namespace> <owner>
```

> Don't change the value of `TENANT_ID`, this value is hardcoded during agent initialization in development mode.

When you open ShellHub UI for the first time, be sure to accept pending device.

See the [devscripts which can be useful for development](./devscripts).

## Authors

ShellHub was created by [O.S. Systems](https://www.ossystems.com.br).

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center"><a href="https://github.com/gustavosbarreto"><img src="https://avatars1.githubusercontent.com/u/86747?v=4?s=100" width="100px;" alt="Luis Gustavo S. Barreto"/><br /><sub><b>Luis Gustavo S. Barreto</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=gustavosbarreto" title="Code">💻</a></td>
      <td align="center"><a href="http://www.ossystems.com.br/blog"><img src="https://avatars0.githubusercontent.com/u/25278?v=4?s=100" width="100px;" alt="Otavio Salvador"/><br /><sub><b>Otavio Salvador</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=otavio" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/leonardojoao"><img src="https://avatars1.githubusercontent.com/u/15831786?v=4?s=100" width="100px;" alt="Leonardo da Rosa Silveira João"/><br /><sub><b>Leonardo da Rosa Silveira João</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=leonardojoao" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/eduardoveiga"><img src="https://avatars3.githubusercontent.com/u/8249343?v=4?s=100" width="100px;" alt="Eduardo Kluwe Veiga"/><br /><sub><b>Eduardo Kluwe Veiga</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=eduardoveiga" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/noreng-jg"><img src="https://avatars2.githubusercontent.com/u/25461720?v=4?s=100" width="100px;" alt="Vagner Nornberg"/><br /><sub><b>Vagner Nornberg</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=noreng-jg" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/fbertux"><img src="https://avatars1.githubusercontent.com/u/2940537?v=4?s=100" width="100px;" alt="Fabio Berton"/><br /><sub><b>Fabio Berton</b></sub></a><br /><a href="#platform-fbertux" title="Packaging/porting to new platform">📦</a></td>
      <td align="center"><a href="http://gomex.me"><img src="https://avatars3.githubusercontent.com/u/95132?v=4?s=100" width="100px;" alt="Rafael Gomes"/><br /><sub><b>Rafael Gomes</b></sub></a><br /><a href="#infra-gomex" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a></td>
    </tr>
    <tr>
      <td align="center"><a href="https://github.com/sixhills"><img src="https://avatars1.githubusercontent.com/u/69159771?v=4?s=100" width="100px;" alt="Mike"/><br /><sub><b>Mike</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=sixhills" title="Code">💻</a></td>
      <td align="center"><a href="https://inductor.me"><img src="https://avatars0.githubusercontent.com/u/20236173?v=4?s=100" width="100px;" alt="inductor(Kohei)"/><br /><sub><b>inductor(Kohei)</b></sub></a><br /><a href="#infra-inductor" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a></td>
      <td align="center"><a href="https://github.com/u5surf"><img src="https://avatars1.githubusercontent.com/u/14180225?v=4?s=100" width="100px;" alt="Y.Horie"/><br /><sub><b>Y.Horie</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=u5surf" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/Robinsondssantos"><img src="https://avatars3.githubusercontent.com/u/29050986?v=4?s=100" width="100px;" alt="Robinson D. S. Santos"/><br /><sub><b>Robinson D. S. Santos</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=Robinsondssantos" title="Code">💻</a></td>
      <td align="center"><a href="https://lbsfilm.at"><img src="https://avatars2.githubusercontent.com/u/1556271?v=4?s=100" width="100px;" alt="Lukas Bachschwell"/><br /><sub><b>Lukas Bachschwell</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=s00500" title="Documentation">📖</a> <a href="https://github.com/shellhub-io/shellhub/commits?author=s00500" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/shawn111"><img src="https://avatars0.githubusercontent.com/u/346761?v=4?s=100" width="100px;" alt="Shawn"/><br /><sub><b>Shawn</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=shawn111" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/brammp"><img src="https://avatars0.githubusercontent.com/u/52255786?v=4?s=100" width="100px;" alt="brammp"/><br /><sub><b>brammp</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=brammp" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center"><a href="http://xevo.com"><img src="https://avatars.githubusercontent.com/u/7035152?v=4?s=100" width="100px;" alt="Sam McKelvie"/><br /><sub><b>Sam McKelvie</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=sammck" title="Code">💻</a></td>
      <td align="center"><a href="http://henrybarreto.dev"><img src="https://avatars.githubusercontent.com/u/23109089?v=4?s=100" width="100px;" alt="Henry Barreto"/><br /><sub><b>Henry Barreto</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=henrybarreto" title="Code">💻</a></td>
      <td align="center"><a href="https://www.linkedin.com/in/antony-rafael-9924511a9/"><img src="https://avatars.githubusercontent.com/u/21010565?v=4?s=100" width="100px;" alt="Antony Rafael"/><br /><sub><b>Antony Rafael</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=AntonyRafael" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/gessecarneiro"><img src="https://avatars.githubusercontent.com/u/55324790?v=4?s=100" width="100px;" alt="Gessé Carneiro"/><br /><sub><b>Gessé Carneiro</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=gessecarneiro" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/benasse"><img src="https://avatars.githubusercontent.com/u/5676729?v=4?s=100" width="100px;" alt="Benoît Stahl"/><br /><sub><b>Benoît Stahl</b></sub></a><br /><a href="https://github.com/shellhub-io/shellhub/commits?author=benasse" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
