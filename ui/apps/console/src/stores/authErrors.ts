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
