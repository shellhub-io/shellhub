export const EXPIRY_OPTIONS = [
  { label: "30 days", value: 30 },
  { label: "60 days", value: 60 },
  { label: "90 days", value: 90 },
  { label: "1 year", value: 365 },
  { label: "Never", value: -1 },
] as const;

export const ROLES = ["administrator", "operator", "observer"] as const;

/** A role that can be assigned to an invitation, member, or API key.
 *  Excludes "owner" — ownership is transferred, not assigned. */
export type AssignableRole = (typeof ROLES)[number];

/** Type guard for strings that happen to be valid AssignableRoles — used to
 *  narrow arbitrary role strings from the backend (e.g. an existing member's
 *  role) before feeding them into RoleSelector. */
export function isAssignableRole(role: unknown): role is AssignableRole {
  return (
    typeof role === "string" &&
    (ROLES as readonly string[]).includes(role)
  );
}

export function isExpired(expiresIn: number): boolean {
  if (expiresIn <= 0) return false;
  return Date.now() > expiresIn * 1000;
}

export function initials(name: string | undefined) {
  if (!name) return "?";
  return name
    .split(/[\s\-_@.]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}
