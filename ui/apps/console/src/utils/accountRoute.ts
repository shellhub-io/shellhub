/**
 * Whether a route is one of the signed-in user's own pages rather than a namespace's.
 */
export function isAccountPath(pathname: string): boolean {
  return pathname === "/profile" || pathname.startsWith("/profile/");
}
