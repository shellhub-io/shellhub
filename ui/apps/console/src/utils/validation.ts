/**
 * Longest accepted display name. Mirrors the server, so the form catches it first.
 */
export const NAME_MAX_LENGTH = 64;

/**
 * Longest accepted username.
 */
export const USERNAME_MAX_LENGTH = 32;
/**
 * The accepted username shape. Lowercase only: usernames are compared case-sensitively by the
 * server, so allowing capitals would let two accounts differ by case alone.
 */
const USERNAME_REGEX = /^[a-z0-9._@-]{3,32}$/;
/**
 * The username rule in words, shown as the field hint and reused as its error — one sentence, so
 * the requirement and the complaint cannot disagree.
 */
export const USERNAME_HINT =
  "3-32 characters: lowercase letters, numbers, hyphens, dots, underscores, @";

/**
 * Shortest accepted password.
 */
const PASSWORD_MIN_LENGTH = 5;
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
const NAMESPACE_NAME_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
/**
 * The namespace-name rule as one sentence, for a field hint.
 */
export const NAMESPACE_NAME_HINT =
  "3-30 characters, lowercase letters, numbers, and hyphens only.";

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
 * A namespace name made from a username: lowercased, every run of other characters turned into
 * one hyphen, cut to the longest accepted name and only then trimmed of hyphens at both ends, so
 * the cut cannot leave one at the end. Returns an empty string when what is left is shorter than
 * the shortest accepted name.
 */
export function suggestNamespaceName(username: string): string {
  const name = username
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .slice(0, NAMESPACE_NAME_MAX_LENGTH)
    .replace(/^-+|-+$/g, "");
  return name.length < NAMESPACE_NAME_MIN_LENGTH ? "" : name;
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

const KEY_NAME_MIN_LENGTH = 3;
/**
 * Longest accepted API or provisioning key name.
 */
export const KEY_NAME_MAX_LENGTH = 20;
/**
 * The key-name rule as one sentence, for a rejection the server reports without naming the part
 * that failed.
 */
export const KEY_NAME_HINT = `Name must be ${KEY_NAME_MIN_LENGTH}–${KEY_NAME_MAX_LENGTH} characters: letters, numbers, - and _ only.`;

/**
 * Checks an API or provisioning key name, returning the message to show or null. Length is
 * reported separately from shape so the error names the rule that was actually broken.
 */
export function validateKeyName(value: string): string | null {
  if (value.length < KEY_NAME_MIN_LENGTH) {
    return `Name must be at least ${KEY_NAME_MIN_LENGTH} characters.`;
  }
  if (value.length > KEY_NAME_MAX_LENGTH) {
    return `Name must be at most ${KEY_NAME_MAX_LENGTH} characters.`;
  }
  if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
    return "Name can only contain letters, numbers, - and _.";
  }
  return null;
}

const GO_FLAG_GROUP = /^[imsU]*(?:-[imsU]*)?[):]/;

/**
 * Whether pattern compiles as a Go regular expression, which is how the server reads firewall and
 * public key patterns. It refuses lookaround and backreferences (`\1` to `\9`, except octal
 * escapes such as `\12`), which Go lacks. It rewrites what only Go accepts, `\Q…\E` quoting, flag
 * groups, `(?P<name>`, repeated anchors, `\p{…}`, `\pL`, `\x{…}` and POSIX classes such as
 * `[:alpha:]`, so the browser's engine can compile the rest.
 *
 * It still refuses a few patterns Go accepts: a class range using a character beyond U+FFFF, a
 * group name that starts with a digit, and a group name used twice. It accepts some that Go
 * refuses, such as unknown escapes or Unicode class names, which the server then rejects with a 400.
 */
export function isGoRegex(pattern: string): boolean {
  let rewritten = "";
  let inClass = false;

  for (let i = 0; i < pattern.length; i++) {
    const char = pattern[i];

    if (char === "\\" && !inClass && pattern[i + 1] === "Q") {
      const end = pattern.indexOf("\\E", i + 2);
      const quoted = pattern.slice(i + 2, end === -1 ? undefined : end);
      rewritten += quoted ? quoted.replace(/[.*+?^${}()|[\]\\]/g, "\\$&") : "(?:)";
      i = end === -1 ? pattern.length : end + 1;
      continue;
    }

    if (char === "\\") {
      const next = pattern[i + 1] ?? "";
      const after = pattern[i + 2] ?? "";
      if (/[pP]/.test(next) || (next === "x" && after === "{")) {
        const end = after === "{" ? pattern.indexOf("}", i + 3) : i + 2;
        if (end === -1) return false;
        rewritten += "\\w";
        i = end;
        continue;
      }
      if (/[89]/.test(next) || (/[1-7]/.test(next) && !/[0-7]/.test(after))) return false;
      rewritten += !inClass && /[bB]/.test(next) ? `(?:\\${next})` : char + next;
      i++;
      continue;
    }

    if (inClass) {
      const posixEnd = char === "[" && pattern[i + 1] === ":" ? pattern.indexOf(":]", i + 2) : -1;
      if (posixEnd !== -1) {
        rewritten += "\\w";
        i = posixEnd + 1;
        continue;
      }
      if (char === "]") inClass = false;
      rewritten += char;
      continue;
    }

    if (char === "[") {
      inClass = true;
      rewritten += "[";
      if (pattern[i + 1] === "^") {
        rewritten += "^";
        i++;
      }
      if (pattern[i + 1] === "]") {
        rewritten += "\\]";
        i++;
      }
      continue;
    }

    if (char === "(" && pattern[i + 1] === "?") {
      const rest = pattern.slice(i + 2);
      if (/^<?[=!]/.test(rest)) return false;
      if (rest.startsWith("P<")) {
        rewritten += "(?<";
        i += 3;
        continue;
      }
      const flags = GO_FLAG_GROUP.exec(rest)?.[0];
      if (flags) {
        rewritten += flags.endsWith(":") ? "(?:" : "(?:)";
        i += 1 + flags.length;
        continue;
      }
    }

    rewritten += char === "^" || char === "$" ? `(?:${char})` : char;
  }

  try {
    new RegExp(rewritten);
    return true;
  } catch {
    return false;
  }
}

/**
 * Checks a firewall rule or public key pattern, returning the message to show or null. label names
 * the field in the message. It checks the trimmed pattern, since that is what the forms send.
 */
export function validatePattern(value: string, label: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return `${label} pattern is required`;
  if (!isGoRegex(trimmed)) return `${label} must be a valid regular expression`;
  return null;
}
