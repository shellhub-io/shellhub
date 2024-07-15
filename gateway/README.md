# Gateway

The `gateway` is a crucial component in ShellHub, managing inbound traffic. It acts as a bridge between
end-users, device agents, and internal services, routing HTTP requests and WebSocket connections to the
appropriate services within the ShellHub Server.

## Features

- **NGINX Process Control**: Manages the lifecycle of the NGINX server process, including starting, stopping, 
and dynamically reloading configurations.
- **Dynamic Configuration**: Generates NGINX configuration files using Go templates, dynamically incorporating
values from environment variables.
- **Certbot Management**: Manages SSL/TLS certificates through Certbot, including the generation, installation,
and renewal of certificates from Let's Encrypt.

## Architecture

The `gateway` is a Go application that utilizes NGINX/OpenResty as a reverse proxy and serves as the entry point
to a Docker image  built from a minimal Alpine Linux base image. On top of this base image, OpenResty is installed
by copying it directly from the official OpenResty Docker image.

## Configuration

Configuration is managed via environment variables. The `gateway` processes NGINX template files found in
the `nginx` directory to create the final NGINX configuration files. These templates are written in Go template
and dynamically incorporate values from the `GatewayConfig` in `config.go`. In development mode, it  watches for
changes in the `nginx` directory, allowing for dynamic updates to NGINX config without needing  service restarts.
