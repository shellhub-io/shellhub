import { validateEmail } from "@/utils/validation";

/**
 * Checks the recovery address. It is optional, so an empty value passes; but it must differ from
 * the primary one, or it would be no recovery at all.
 */
export function validateRecoveryEmail(
  recoveryEmail: string,
  primaryEmail: string,
): string | null {
  if (!recoveryEmail) return null;
  const formatError = validateEmail(recoveryEmail);
  if (formatError) return formatError;
  if (recoveryEmail.toLowerCase() === primaryEmail.toLowerCase())
    return "Must be different from your email";
  return null;
}
