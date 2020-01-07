# ShellHub

ShellHub is a modern SSH server for remotely accessing Linux devices
via command line (using any SSH client) or web-based user interface.
It is intended to be used instead of sshd. ShellHub enables teams to
easily access any Linux device behind firewall and NAT.

## Installing and Running

ShellHub is built using the microservices design pattern, meaning that
multiple small, isolated services make up the server.
In order to make it easy to test ShellHub as a whole, we have created
a Docker Compose environment that brings all of these components up
and connects them together on a single machine.

> Make sure you have installed docker and docker-compose

In a working directory, download the docker-compose file:

```
$ wget https://raw.githubusercontent.com/shellhub-io/shellhub/master/docker-compose.yml
```

Brings up the ShellHub server:

```
$ docker-compose up
```

### Creating the initial user

After ShellHub server have been up and running you need to create the initial user.

Go to the working directory and download the `add-user` utility from our repository:

```
$ wget https://raw.githubusercontent.com/shellhub-io/shellhub/master/bin/add-user
```

Then run the `add-user` utility:

```
$ ./add-user <username> <password>
```

### Open the ShellHub UI

The ShellHub UI can now be accessed at http://localhost

## Development

```
$ docker-compose build
$ docker-compose up
```

