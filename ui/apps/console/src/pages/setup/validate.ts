import { NAMESPACE_NAME_MAX_LENGTH } from "@/utils/validation";

const USERNAME_REGEX = /^[a-z0-9\-_.@]+$/;

// suggestNamespace derives a namespace name from the username, so setup can pre-fill it
// (readonly) and let the user override it. Lowercases, turns runs of invalid characters into a
// single hyphen, trims leading/trailing hyphens, and caps at the namespace max length. The
// result is validated by the shared validateNamespaceName (see setupResolver).
export function suggestNamespace(username: string): string {
  return username
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-+/, "")
    .slice(0, NAMESPACE_NAME_MAX_LENGTH)
    .replace(/-+$/, "");
}

export interface FormErrors {
  name?: string;
  username?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

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
