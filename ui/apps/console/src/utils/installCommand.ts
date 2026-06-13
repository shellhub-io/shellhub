/**
 * Builds the ShellHub agent install command. Pass the credential env pair:
 * `TENANT_ID=<id>` to land the device in a namespace's pending list, or
 * `PAIRING_CODE=<code>` for a pre-authorized install that is accepted
 * automatically and can be confirmed live.
 */
export function buildInstallCommand(
  credential: string,
  serverAddress: string,
): string {
  return [
    `curl -sSf ${serverAddress}/install.sh | \\`,
    `        ${credential} \\`,
    `        SERVER_ADDRESS=${serverAddress} \\`,
    "        sh",
  ].join("\n");
}
