/**
 * Whether a route belongs to the admin console rather than to a namespace.
 */
export function isAdminPath(pathname: string): boolean {
  return pathname === "/admin" || pathname.startsWith("/admin/");
}

/**
 * Where the admin console sends someone who is not an instance admin. It belongs to no context,
 * so it opens no tab of its own.
 */
export const ADMIN_UNAUTHORIZED_PATH = "/admin/unauthorized";
