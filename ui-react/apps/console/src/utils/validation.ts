export function validatePassword(value: string): string | null {
  if (value.length < 5 || value.length > 32)
    return "Password must be 5–32 characters long";
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

export function isMaxNamespacesValid(
  limitEnabled: boolean,
  limitDisabled: boolean,
  maxNamespaces: string,
): boolean {
  return !limitEnabled || limitDisabled || parseInt(maxNamespaces, 10) >= 1;
}
