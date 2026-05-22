export const NAME_MIN_LENGTH = 1;
export const NAME_MAX_LENGTH = 64;

export const USERNAME_MIN_LENGTH = 3;
export const USERNAME_MAX_LENGTH = 32;
export const USERNAME_REGEX = /^[a-z0-9._@-]{3,32}$/;
export const USERNAME_HINT =
  "3-32 characters: lowercase letters, numbers, hyphens, dots, underscores, @";

export const PASSWORD_MIN_LENGTH = 5;
export const PASSWORD_MAX_LENGTH = 32;
export const PASSWORD_HINT = "5-32 characters";

export const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function validateName(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Name is required";
  if (trimmed.length > NAME_MAX_LENGTH)
    return `Name must be at most ${NAME_MAX_LENGTH} characters`;
  return null;
}

export function validateUsername(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Username is required";
  if (!USERNAME_REGEX.test(trimmed)) return USERNAME_HINT;
  return null;
}

export function validateEmail(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Email is required";
  if (!EMAIL_REGEX.test(trimmed)) return "Enter a valid email address";
  return null;
}

export function validatePassword(value: string): string | null {
  if (value.length < PASSWORD_MIN_LENGTH || value.length > PASSWORD_MAX_LENGTH)
    return `Password must be ${PASSWORD_MIN_LENGTH}–${PASSWORD_MAX_LENGTH} characters long`;
  return null;
}

export const NAMESPACE_NAME_MIN_LENGTH = 3;
export const NAMESPACE_NAME_MAX_LENGTH = 30;
export const NAMESPACE_NAME_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
export const NAMESPACE_NAME_HINT =
  "3-30 characters, lowercase letters, numbers, and hyphens only.";

export const NAMESPACE_NAME_RULES: readonly string[] = [
  "3-30 characters",
  "Lowercase letters, numbers, and hyphens only",
  "Cannot begin or end with a hyphen",
];

export function validateNamespaceName(name: string): string | null {
  if (name.length < NAMESPACE_NAME_MIN_LENGTH) {
    return `Name must be at least ${NAMESPACE_NAME_MIN_LENGTH} characters`;
  }
  if (name.length > NAMESPACE_NAME_MAX_LENGTH) {
    return `Name must be at most ${NAMESPACE_NAME_MAX_LENGTH} characters`;
  }
  if (!NAMESPACE_NAME_REGEX.test(name)) {
    return "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)";
  }
  return null;
}

export const MAX_NAMESPACES_ERROR =
  "Max namespaces must be a number greater than or equal to 1";

export function isMaxNamespacesValid(
  limitEnabled: boolean,
  limitDisabled: boolean,
  maxNamespaces: string,
): boolean {
  return !limitEnabled || limitDisabled || parseInt(maxNamespaces, 10) >= 1;
}
