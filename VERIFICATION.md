# Verifying Container Images

All ShellHub container images published to Docker Hub are signed using [cosign](https://docs.sigstore.dev/cosign/overview/) with keyless (OIDC) signing via [Sigstore](https://www.sigstore.dev/). Signatures are generated automatically during CI/CD using GitHub Actions' OIDC identity and recorded in the [Rekor](https://docs.sigstore.dev/logging/overview/) transparency log.

## Prerequisites

Install cosign: https://docs.sigstore.dev/cosign/system_config/installation/

## Verifying an image

```bash
cosign verify \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  shellhubio/<image>:<tag>
```

Replace `<image>` with the component name (`api`, `ssh`, `gateway`, `ui`, `cli`, `agent`) and `<tag>` with the version (e.g., `v0.26.0`).

### Example

```bash
cosign verify \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  shellhubio/api:v0.18.0
```

A successful verification prints the signature payload and confirms that:

- The image has not been modified since it was published
- The image was built by a GitHub Actions workflow in the `shellhub-io/shellhub` repository
- The signing event is recorded in a public transparency log
