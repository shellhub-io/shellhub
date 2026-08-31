const USERNAME_REGEX = /^[a-z0-9_.@-]+$/;

/**
 * Checks a display name, returning the message to show or null.
 */
export function validateName(v: string): string | null {
  if (!v.trim()) return "Name is required";
  if (v.length > 64) return "Name must be at most 64 characters";
  return null;
}

/**
 * Checks a username. Lowercase is enforced here rather than silently applied, so the user sees
 * the name they will actually have.
 */
export function validateUsername(v: string): string | null {
  if (!v.trim()) return "Username is required";
  if (v.length > 32) return "Username must be at most 32 characters";
  if (v !== v.toLowerCase()) return "Username must be lowercase";
  if (v.includes(" ")) return "Username cannot contain spaces";
  if (!USERNAME_REGEX.test(v))
    return "Only lowercase letters, numbers, dots, underscores, @ and hyphens are allowed";
  return null;
}

/**
 * Checks an email address.
 */
export function validateEmail(v: string): string | null {
  if (!v.trim()) return "Email is required";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return "Invalid email format";
  return null;
}

/**
 * Checks the recovery address. It is optional, so an empty value passes; but it must differ from
 * the primary one, or it would be no recovery at all.
 */
export function validateRecoveryEmail(recoveryEmail: string, primaryEmail: string): string | null {
  if (!recoveryEmail) return null;
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(recoveryEmail)) return "Invalid email format";
  if (recoveryEmail.toLowerCase() === primaryEmail.toLowerCase()) return "Must be different from your email";
  return null;
}
