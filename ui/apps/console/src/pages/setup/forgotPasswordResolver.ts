import type { FieldErrors, Resolver } from "react-hook-form";
import { validateIdentifier } from "@/utils/validation";

/**
 * The forgot-password form: one field taking either a username or an email.
 */
export interface ForgotPasswordFormValues {
  account: string;
}

/**
 * Validates the forgot-password form for react-hook-form.
 */
export const forgotPasswordResolver: Resolver<ForgotPasswordFormValues> = (values) => {
  const errors: FieldErrors<ForgotPasswordFormValues> = {};

  const accountError = validateIdentifier(values.account);
  if (accountError) {
    errors.account = { type: "validate", message: accountError };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values: { ...values, account: values.account.trim() }, errors: {} };
};
