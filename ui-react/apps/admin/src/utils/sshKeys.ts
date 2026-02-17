const SSH_KEY_PREFIXES = [
  "ssh-rsa",
  "ssh-dss",
  "ssh-ed25519",
  "ecdsa-sha2-nistp256",
  "ecdsa-sha2-nistp384",
  "ecdsa-sha2-nistp521",
];

const PEM_BEGIN = "-----BEGIN";

/**
 * Basic browser-side validation for SSH public keys.
 * Checks for OpenSSH format (ssh-rsa ...) or PEM format (-----BEGIN ...).
 */
export function isPublicKeyValid(key: string): boolean {
  const trimmed = key.trim();
  if (!trimmed) return false;

  // OpenSSH format: "<type> <base64-data> [comment]"
  for (const prefix of SSH_KEY_PREFIXES) {
    if (trimmed.startsWith(prefix)) {
      const parts = trimmed.split(/\s+/);
      if (parts.length >= 2 && parts[1].length > 10) return true;
    }
  }

  // PEM format
  if (trimmed.startsWith(PEM_BEGIN)) {
    return trimmed.includes("-----END");
  }

  return false;
}
