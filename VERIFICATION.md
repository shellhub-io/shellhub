# Verifying Releases

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

Replace `<image>` with the component name (`server`, `gateway`, `ui`, `agent`) and `<tag>` with the version (e.g., `v0.26.0`).

### Example

```bash
cosign verify \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  shellhubio/server:v0.18.0
```

A successful verification prints the signature payload and confirms that:

- The image has not been modified since it was published
- The image was built by a GitHub Actions workflow in the `shellhub-io/shellhub` repository
- The signing event is recorded in a public transparency log

## Verifying SBOM attestations

Each image also carries a signed [CycloneDX](https://cyclonedx.org/) SBOM wrapped in an [in-toto](https://in-toto.io/) attestation. The attestation is signed with the same keyless flow as the image itself, so verifying it proves the SBOM has not been tampered with since it was generated during the release build.

```bash
cosign verify-attestation \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --type cyclonedx \
  shellhubio/<image>:<tag>
```

### Example

```bash
cosign verify-attestation \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --type cyclonedx \
  shellhubio/server:v0.18.0
```

To extract the SBOM payload from the verified attestation:

```bash
cosign verify-attestation \
  --certificate-identity-regexp="https://github.com/shellhub-io/shellhub/" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --type cyclonedx \
  shellhubio/<image>:<tag> | jq -r '.payload' | base64 -d | jq '.predicate'
```

## Verifying agent binaries

Each release publishes `checksums.txt` with the SHA-256 checksum of every file attached to it. The install script checks the agent binary against it before installing.

The release also publishes `checksums.txt.sigstore.json`, a cosign bundle that signs `checksums.txt` with the same keyless flow as the images. To check the signature yourself, download both files and the binary from the release and run:

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp="^https://github\.com/shellhub-io/shellhub/\.github/workflows/build-agent\.yml@refs/tags/v" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  checksums.txt
grep " shellhub-agent-linux-amd64.gz$" checksums.txt | sha256sum -c
```

`cosign verify-blob` prints `Verified OK` when the release workflow of a ShellHub tag signed `checksums.txt`. `sha256sum` then prints `OK` when the binary matches it. Replace `amd64` with the architecture you downloaded.
