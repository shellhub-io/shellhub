import type { FieldErrors, Resolver } from "react-hook-form";
import { validateIdentifier } from "@/utils/validation";

/**
 * The sign-in form's fields.
 */
export interface LoginFormValues {
  username: string;
  password: string;
}

/**
 * Validates the sign-in form. The identifier may be a username or an email, so it is checked
 * against both rather than either.
 */
export const loginResolver: Resolver<LoginFormValues> = (values) => {
  const errors: FieldErrors<LoginFormValues> = {};

  const usernameError = validateIdentifier(values.username);
  if (usernameError) {
    errors.username = { type: "validate", message: usernameError };
  }

  if (!values.password) {
    errors.password = { type: "required", message: "Password is required" };
  }

  if (Object.keys(errors).length > 0) {
    return { values: {}, errors };
  }

  return { values: { ...values, username: values.username.trim() }, errors: {} };
};
