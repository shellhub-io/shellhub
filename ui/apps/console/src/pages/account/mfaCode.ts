import { apiErrorFields } from "@/api/errors";

/**
 * What the form says when an account with MFA leaves the code blank on a change that needs one.
 */
export const MFA_CODE_REQUIRED_MESSAGE =
  "Enter a code from your authenticator app or a recovery code";

/**
 * The hint for any field that accepts an MFA code, warning that a recovery code entered there is
 * used up.
 */
export const MFA_CODE_HINT =
  "The 6-digit code from your authenticator app, or a recovery code, which is then used up.";

const MFA_CODE_REFUSAL_MESSAGES: Record<string, string> = {
  "is invalid": "The code is incorrect.",
  "too many invalid codes, try again later":
    "Too many incorrect codes. Try again later.",
};

function mfaCodeRefusalMessage(reason: string): string {
  return MFA_CODE_REFUSAL_MESSAGES[reason] ?? MFA_CODE_REQUIRED_MESSAGE;
}

/**
 * The message to put beside the MFA code field when err is the server refusing the code, or
 * undefined when err is about something else.
 */
export function mfaCodeRefusal(err: unknown): string | undefined {
  const reason = apiErrorFields(err).mfa_code;
  return reason ? mfaCodeRefusalMessage(reason) : undefined;
}
