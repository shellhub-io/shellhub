import { rolePermissions, type Permission } from "./permissions.generated";

export type { Permission };

/**
 * A namespace role, one of the roles the Go authorizer can assign. Generated with the permission
 * table, so a role the server adds or drops changes this type too.
 */
export type Role = keyof typeof rolePermissions;

/**
 * Narrows an arbitrary string — a stored value, an API field — to a Role. Anything unrecognised
 * is not a role, which is what keeps a renamed or absent role from being treated as an owner.
 */
export const isValidRole = (role: string | null): role is Role =>
  role !== null && Object.prototype.hasOwnProperty.call(rolePermissions, role);

/**
 * Whether a role holds every permission given, the Go authorizer's own check for the routes the
 * UI calls. A role that is not valid holds none. This decides what the UI offers, not what the
 * API permits: the server checks again, and this only keeps a control from being shown that
 * would fail.
 */
export function hasPermission(
  role: string | null,
  required: Permission | readonly Permission[],
): boolean {
  if (!isValidRole(role)) return false;
  const held: readonly Permission[] = rolePermissions[role];
  const wanted: readonly Permission[] = typeof required === "string" ? [required] : required;
  return wanted.every((permission) => held.includes(permission));
}
