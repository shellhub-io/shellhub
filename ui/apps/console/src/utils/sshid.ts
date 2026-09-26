/**
 * Builds the SSHID for a device.
 * Format: <namespace>.<deviceName>@<server>
 * Matches the Vue UI formula: `${item.namespace}.${item.name}@${window.location.hostname}`
 */
export function buildSshid(namespace: string, deviceName: string): string {
  return `${namespace}.${deviceName}@${window.location.hostname}`;
}

/**
 * The parts of an SSHID: the namespace, the device (which may itself hold dots) and the server.
 * Returns null for a value without a server part, such as the bare device UID a caller falls back
 * to before the namespace name has loaded.
 */
export function parseSshid(
  sshid: string,
): { namespace: string; device: string; host: string } | null {
  const at = sshid.lastIndexOf("@");
  if (at < 0) return null;
  const target = sshid.slice(0, at);
  const dot = target.indexOf(".");
  return {
    namespace: dot < 0 ? "" : target.slice(0, dot),
    device: target.slice(dot + 1),
    host: sshid.slice(at + 1),
  };
}

/**
 * The login a connection uses when the user leaves it empty.
 */
export const DEFAULT_LOGIN = "root";

/**
 * The ssh:// URL that hands a login on the SSHID's device to the user's own SSH client. The
 * userinfo is "login@namespace.device", percent-encoded because it carries an @. Returns null when
 * the SSHID has no server part to connect to.
 */
export function sshUrl(sshid: string, login: string): string | null {
  const parts = parseSshid(sshid);
  if (!parts) return null;
  const target = parts.namespace
    ? `${parts.namespace}.${parts.device}`
    : parts.device;
  return `ssh://${encodeURIComponent(`${login}@${target}`)}@${parts.host}`;
}
