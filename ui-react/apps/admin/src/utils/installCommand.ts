/**
 * Builds the ShellHub agent install command.
 * Used in WelcomeScreen and WizardStep2Install — single source of truth.
 */
export function buildInstallCommand(
  tenantId: string,
  serverAddress: string,
): string {
  return [
    `curl -sSf ${serverAddress}/install.sh | \\`,
    `        TENANT_ID=${tenantId} \\`,
    `        SERVER_ADDRESS=${serverAddress} \\`,
    `        sh`,
  ].join("\n");
}
