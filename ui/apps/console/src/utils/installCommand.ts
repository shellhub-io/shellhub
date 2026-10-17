/**
 * Builds the ShellHub agent install command. Pass the credential env pair:
 * `PROVISIONING_KEY=<key>` to enroll into the key's namespace under the key's mode,
 * or `TENANT_ID=<id>` to land the device in a legacy namespace's pending list.
 */
export function buildInstallCommand(
  credential: string,
  serverAddress: string,
): string {
  return [
    `curl -sSf ${serverAddress}/install.sh | \\`,
    `        ${credential} \\`,
    "        sh",
  ].join("\n");
}

/**
 * Builds the one-line Docker connector install command. Pass the same credential env pair as
 * {@link buildInstallCommand}: every container the connector finds enrolls with it.
 */
export function buildConnectorCommand(
  credential: string,
  serverAddress: string,
): string {
  return `curl -sSf ${serverAddress}/install.sh | ${credential} SERVER_ADDRESS=${serverAddress} sh -s connector`;
}
