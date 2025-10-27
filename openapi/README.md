<p align="center">
  <img src="assets/logo.png" alt="shellhub OpenAPI spec" />
</p>

<p align="center"><strong>ShellHub OpenAPI for Community, Cloud and Enterprise</strong></p>

<p align="center">
<img src="https://img.shields.io/badge/openapi-6BA539?style=for-the-badge&logo=OpenAPI Initiative&logoColor=white" />
<img src="https://img.shields.io/badge/swagger-85EA2D?style=for-the-badge&logo=Swagger&logoColor=black" />
</p>

## How it works?

It is simples; When ShellHub is in developement mode (with the `SHELLHUB_ENV`
set to `development`), it will enable, through the `gateway`, a [URL to preview the
OpenAPI](http://localhost/openapi/) in any OpenAPI file after a page reload.

## How to use?

### Preview

Its usage is simple too, you just need `up` the `shellhub-io/shellhub` containers in development
mode and access the URL.

- Access the URL [http://localhost/openapi/preview](http://localhost/openapi) to check the preview.

### Lint

You can also linter the OpenAPI instances to check if everything is correct. You just need to run
the script called `lint` inside `scripts` folder.

```bash
Lint the OpenAPI instances.

Usage:
    ./scripts/lint <instance>

Instances:
  community   Lint the community instance
  cloud       Lint the cloud instance
  enterprise  Lint the enterprise instance

Options:
    --help  Display this help message
```

### Mock

If you need to mock a ShellHub instance, the `scripts` folder has another one called `mock`,
what will to make available a full service to you test the API based on instance you chose in
[http://localhost/openapi/mock](http://localhost/openapi/mock).

```bash
Mock the OpenAPI instances.

Usage:
    ./scripts/mock <instance>

Instances:
  community   Mock the community instance
  cloud       Mock the cloud instance
  enterprise  Mock the enterprise instance

Options:
    --help  Display this help message
```

### Proxy

If you need to check if the API contract is right, you can use the script called `proxy`,
what will to make available a full service to you send request to the API based on instance
you chose in [http://localhost/openapi/proxy](http://localhost/openapi/proxy).

```bash
Proxy the OpenAPI instances.

Usage:
    ./scripts/proxy <instance>

Instances:
  community   Proxy the community instance
  cloud       Proxy the cloud instance
  enterprise  Proxy the enterprise instance

Options:
    --help  Display this help message

```

### Generate

You can generate a TypeScript client from the OpenAPI specification for any of the available instances using the
generate script. This will build the client and place it in the specified output directory.

```bash
Generate the TypeScript client.

Usage:
    ./scripts/generate <output-directory>

Options:
    --help  Display this help message
```

For example, to generate a client in the `../shellhub/ui/src/api/client/` directory, you would run:

```bash
./scripts/generate ../shellhub/ui/src/api/client/
```

The TypeScript client will be generated in the specified folder and can be used to interact with the ShellHub API.

## Integrations

### Community

Today, the Community has a script called `openapi` to execute commands against the OpenAPI
spec provided by ShellHub `gateway` and this repository.

<p align="center">. . .</p>
