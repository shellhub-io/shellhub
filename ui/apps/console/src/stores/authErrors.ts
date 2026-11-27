import { isSdkError } from "@/api/errors";

/**
 * mfaResetCodesRejectedMessage is shown when completing a reset fails for any reason, a refused
 * code pair or otherwise. The server cancels the codes after three pairs with one right code, so
 * the text points the user to a new reset.
 */
export const mfaResetCodesRejectedMessage =
  "Invalid verification codes. Three wrong attempts cancel the codes, so request a new reset if this keeps failing.";

/**
 * mfaResetRequestError turns a failed request for MFA reset emails into the text to show: a 429
 * says one went out within the minute, anything else says the emails could not be sent.
 */
export function mfaResetRequestError(err: unknown): string {
  if (isSdkError(err) && err.status === 429) {
    return "A reset was already requested in the last minute. Check both inboxes, or wait a minute and try again.";
  }

  return "Unable to send reset emails. Please try again.";
}

/**
 * MfaTokenMissingError is thrown by a second-factor sign-in action called without the MFA token
 * the password step leaves.
 */
export class MfaTokenMissingError extends Error {
  override name = "MfaTokenMissingError";
}

/**
 * mfaCodeErrorMessage turns a failed authenticator-code sign-in into the text to show: a 429 says
 * to wait, anything else says the code is invalid. It does not single out MfaTokenMissingError,
 * since the code form is never shown without the password step's MFA token.
 */
export function mfaCodeErrorMessage(err: unknown): string {
  if (isSdkError(err) && err.status === 429) {
    return "Too many failed attempts. Please wait before trying again.";
  }

  return "Invalid verification code";
}

/**
 * recoveryCodeErrorMessage turns a failed recovery-code sign-in into the text to show: without the
 * password step's MFA token it asks for the password first, anything else says the code is
 * invalid.
 */
export function recoveryCodeErrorMessage(err: unknown): string {
  if (err instanceof MfaTokenMissingError) return "Sign in with your password first.";

  return "Invalid recovery code";
}
