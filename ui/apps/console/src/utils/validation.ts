/**
 * Shortest accepted display name.
 */
export const NAME_MIN_LENGTH = 1;
/**
 * Longest accepted display name. Mirrors the server, so the form catches it first.
 */
export const NAME_MAX_LENGTH = 64;

/**
 * Shortest accepted username.
 */
export const USERNAME_MIN_LENGTH = 3;
/**
 * Longest accepted username.
 */
export const USERNAME_MAX_LENGTH = 32;
/**
 * The accepted username shape. Lowercase only: usernames are compared case-sensitively by the
 * server, so allowing capitals would let two accounts differ by case alone.
 */
export const USERNAME_REGEX = /^[a-z0-9._@-]{3,32}$/;
/**
 * The username rule in words, shown as the field hint and reused as its error — one sentence, so
 * the requirement and the complaint cannot disagree.
 */
export const USERNAME_HINT =
  "3-32 characters: lowercase letters, numbers, hyphens, dots, underscores, @";

/**
 * Shortest accepted password.
 */
export const PASSWORD_MIN_LENGTH = 5;
/**
 * Longest accepted password.
 */
export const PASSWORD_MAX_LENGTH = 32;
/**
 * The password length rule in words, shown as the field hint.
 */
export const PASSWORD_HINT = "5-32 characters";

/**
 * A deliberately permissive email shape. Real validation is the confirmation mail; this only
 * catches an address that could not be one at all, and must not reject a valid odd address.
 */
export const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Checks a display name, returning the message to show or null. Trims first, so whitespace alone
 * is empty rather than valid.
 */
export function validateName(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Name is required";
  if (trimmed.length > NAME_MAX_LENGTH)
    return `Name must be at most ${NAME_MAX_LENGTH} characters`;
  return null;
}

/**
 * Checks a username, returning the message to show or null.
 */
export function validateUsername(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Username is required";
  if (!USERNAME_REGEX.test(trimmed)) return USERNAME_HINT;
  return null;
}

/**
 * Checks an email address, returning the message to show or null.
 */
export function validateEmail(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Email is required";
  if (!EMAIL_REGEX.test(trimmed)) return "Enter a valid email address";
  return null;
}

/**
 * Checks a sign-in identifier, which may be either a username or an email — the login form takes
 * one field for both, so neither rule alone can reject it.
 */
export function validateIdentifier(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return "Username or email is required";
  if (EMAIL_REGEX.test(trimmed) || USERNAME_REGEX.test(trimmed)) return null;
  return "Enter a valid username or email";
}

/**
 * Checks a password's length. Not trimmed: leading and trailing spaces are part of a password.
 */
export function validatePassword(value: string): string | null {
  if (value.length < PASSWORD_MIN_LENGTH || value.length > PASSWORD_MAX_LENGTH)
    return `Password must be ${PASSWORD_MIN_LENGTH}–${PASSWORD_MAX_LENGTH} characters long`;
  return null;
}

/**
 * Shortest accepted namespace name.
 */
export const NAMESPACE_NAME_MIN_LENGTH = 3;
/**
 * Longest accepted namespace name.
 */
export const NAMESPACE_NAME_MAX_LENGTH = 30;
/**
 * The accepted namespace-name shape. A namespace name appears in an SSHID and in a hostname, so
 * it is restricted to what is legal there — and cannot begin or end with a hyphen.
 */
export const NAMESPACE_NAME_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
/**
 * The namespace-name rule as one sentence, for a field hint.
 */
export const NAMESPACE_NAME_HINT =
  "3-30 characters, lowercase letters, numbers, and hyphens only.";

/**
 * The same rule broken into checklist lines, for the create form, where the requirements are
 * shown as a list the user watches turn green.
 */
export const NAMESPACE_NAME_RULES: readonly string[] = [
  "3-30 characters",
  "Lowercase letters, numbers, and hyphens only",
  "Cannot begin or end with a hyphen",
];

/**
 * Checks a namespace name, returning the message to show or null. Length is reported separately
 * from shape so the error names the rule that was actually broken.
 */
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

/**
 * The message shown when the namespace limit is not a usable number.
 */
export const MAX_NAMESPACES_ERROR =
  "Max namespaces must be a number greater than or equal to 1";

/**
 * Whether the namespace-limit field holds an acceptable value. A limit that is switched off, or
 * not enabled at all, is valid whatever the field says — only an enforced limit has to parse
 * as at least one.
 */
export function isMaxNamespacesValid(
  limitEnabled: boolean,
  limitDisabled: boolean,
  maxNamespaces: string,
): boolean {
  return !limitEnabled || limitDisabled || parseInt(maxNamespaces, 10) >= 1;
}
