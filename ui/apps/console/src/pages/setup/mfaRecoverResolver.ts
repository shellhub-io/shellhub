import type { FieldErrors, Resolver } from "react-hook-form";

export interface MfaRecoverFormValues {
  recoveryCode: string;
}

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
