import { NAMESPACE_NAME_MAX_LENGTH } from "@/utils/validation";

const USERNAME_REGEX = /^[a-z0-9\-_.@]+$/;

/**
 * Derives a namespace name from a username, so setup can pre-fill one. Lowercases, collapses
 * runs of invalid characters into a single hyphen, trims hyphens from the ends and caps the
 * length. A suggestion only — the field stays editable, and the result is validated by the
 * shared validateNamespaceName like any other.
 */
export function suggestNamespace(username: string): string {
  return username
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-+/, "")
    .slice(0, NAMESPACE_NAME_MAX_LENGTH)
    .replace(/-+$/, "");
}

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
 * Checks the account fields shared by setup, sign-up and invitation, so the three cannot come to
 * disagree about what a valid username or password is.
 */
export function validate(fields: {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}): FormErrors {
  const errors: FormErrors = {};

  if (fields.name.length < 1 || fields.name.length > 64) {
    errors.name = "Name must be 1-64 characters long";
  }

  if (fields.username.length < 3 || fields.username.length > 32) {
    errors.username = "Username must be 3-32 characters long";
  } else if (!USERNAME_REGEX.test(fields.username)) {
    errors.username = "Only lowercase letters, numbers, and -_.@ are allowed";
  }

  if (!fields.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
    errors.email = "Enter a valid email address";
  }

  if (fields.password.length < 5 || fields.password.length > 32) {
    errors.password = "Password must be 5-32 characters long";
  }

  if (fields.confirmPassword !== fields.password) {
    errors.confirmPassword = "Passwords do not match";
  }

  return errors;
}
