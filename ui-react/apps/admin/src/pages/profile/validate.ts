export function validateRecoveryEmail(recoveryEmail: string, primaryEmail: string): string | null {
  if (!recoveryEmail) return null;
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(recoveryEmail)) return "Invalid email format";
  if (recoveryEmail.toLowerCase() === primaryEmail.toLowerCase()) return "Must be different from your email";
  return null;
}
