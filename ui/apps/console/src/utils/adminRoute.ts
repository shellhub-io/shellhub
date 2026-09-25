/**
 * Whether a route belongs to the admin console rather than to a namespace.
 */
export function isAdminPath(pathname: string): boolean {
  return pathname === "/admin" || pathname.startsWith("/admin/");
}
