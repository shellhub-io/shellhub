import type { FieldErrors, Resolver } from "react-hook-form";

/**
 * The MFA recovery form's fields.
 */
export interface MfaRecoverFormValues {
  recoveryCode: string;
}

/**
 * Validates a recovery code, trimming it first — these are copied from a saved file, and often
 * arrive with whitespace attached.
 */
export const mfaRecoverResolver: Resolver<MfaRecoverFormValues> = (values) => {
  const recoveryCode = values.recoveryCode.trim();
  const errors: FieldErrors<MfaRecoverFormValues> = {};

  if (!recoveryCode) {
    errors.recoveryCode = {
      type: "required",
      message: "Recovery code is required",
    };
    return { values: {}, errors };
  }

  return { values: { ...values, recoveryCode }, errors: {} };
};
