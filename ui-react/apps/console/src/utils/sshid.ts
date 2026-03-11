/**
 * Builds the SSHID for a device.
 * Format: <namespace>.<deviceName>@<server>
 * Matches the Vue UI formula: `${item.namespace}.${item.name}@${window.location.hostname}`
 */
export function buildSshid(namespace: string, deviceName: string): string {
  return `${namespace}.${deviceName}@${window.location.hostname}`;
}
