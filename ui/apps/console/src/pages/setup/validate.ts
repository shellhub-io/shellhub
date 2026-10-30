import {
  validateEmail,
  validateName,
  validatePassword,
  validateUsername,
} from "@/utils/validation";

/**
 * The per-field errors of the account forms, keyed by field so each lands on its own input.
 */
export interface FormErrors {
  name?: string;
  username?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

/**
 * Checks the account fields shared by setup, sign-up and invitation against the same rules as
 * every other account form.
 */
export function validate(fields: {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}): FormErrors {
  const errors: FormErrors = {
    name: validateName(fields.name) ?? undefined,
    username: validateUsername(fields.username) ?? undefined,
    email: validateEmail(fields.email) ?? undefined,
    password: validatePassword(fields.password) ?? undefined,
  };

  if (fields.confirmPassword !== fields.password) {
    errors.confirmPassword = "Passwords do not match";
  }

  return errors;
}
