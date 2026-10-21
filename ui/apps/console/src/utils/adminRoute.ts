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

/**
 * The admin console's landing page, the instance overview.
 */
export const ADMIN_INSTANCE_PATH = "/admin/instance";

/**
 * The instance's sign-in settings page. Only enterprise has it; any other edition redirects it to
 * the instance overview.
 */
export const ADMIN_AUTHENTICATION_PATH = "/admin/settings/authentication";

/**
 * The instance's telemetry setting and payload preview, on every edition with an admin console.
 */
export const ADMIN_TELEMETRY_PATH = "/admin/settings/telemetry";
