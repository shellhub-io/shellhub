/**
 * The route of this browser's preferences, which belong to no namespace: they are reached without
 * one, and kept in the browser rather than in the account.
 */
export const PREFERENCES_PATH = "/preferences";

/**
 * Whether a route is one of this browser's preferences rather than a namespace's or the account's
 * page.
 */
export function isPreferencesPath(pathname: string): boolean {
  return (
    pathname === PREFERENCES_PATH || pathname.startsWith(`${PREFERENCES_PATH}/`)
  );
}
