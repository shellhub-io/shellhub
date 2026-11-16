/**
 * Where a device enrolment code waits while the user signs in. It outlives the page, so it has
 * to be storage rather than component state.
 */
export const PENDING_DEVICE_CODE_KEY = "shellhub:pending-device-code";

/**
 * Remembers an enrolment code across a sign-in, for a user who followed an accept-device link
 * while signed out.
 */
export function setPendingDeviceCode(code: string): void {
  localStorage.setItem(PENDING_DEVICE_CODE_KEY, code);
}

/**
 * Forgets any pending enrolment code.
 */
export function clearPendingDeviceCode(): void {
  localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
}

/**
 * Whether an enrolment code is waiting to be used.
 */
export function hasPendingDeviceCode(): boolean {
  return getPendingDeviceCode() !== null;
}

/**
 * Reads the pending enrolment code without clearing it. Returns null when there is none.
 */
export function getPendingDeviceCode(): string | null {
  return localStorage.getItem(PENDING_DEVICE_CODE_KEY);
}

/**
 * Reads the pending enrolment code and clears it in the same step, so a code is acted on once.
 * Returns null when there is none.
 */
export function consumePendingDeviceCode(): string | null {
  const code = getPendingDeviceCode();
  if (code) clearPendingDeviceCode();
  return code;
}

/**
 * Where a first SSO sign-in that matched an existing account waits while the user signs in to
 * that account by its own credentials. The link it waits for belongs to this browser, so the tab
 * that landed on it is the one that confirms it.
 */
export const PENDING_SSO_LINK_KEY = "shellhub:pending-sso-link";

/**
 * Remembers that the next sign-in in this tab confirms an SSO link.
 */
export function setPendingSsoLink(): void {
  sessionStorage.setItem(PENDING_SSO_LINK_KEY, "1");
}

/**
 * Forgets a pending SSO link, for a sign-in that no longer confirms one.
 */
export function clearPendingSsoLink(): void {
  sessionStorage.removeItem(PENDING_SSO_LINK_KEY);
}

/**
 * Reads the redirect query parameter, returning it only if it is a path within this app.
 * Anything else — an absolute URL, a protocol-relative //host, or a backslash the browser may
 * normalize into one — falls back, because a redirect a caller controls is an open redirect.
 */
export function getSafeRedirect(
  params: URLSearchParams,
  fallback = "/dashboard",
): string {
  const raw = params.get("redirect");
  if (
    raw &&
    raw.startsWith("/") &&
    !raw.startsWith("//") &&
    !raw.startsWith("/\\")
  ) {
    return raw;
  }
  return fallback;
}

/**
 * Where to land after a sign-in. A pending SSO link wins and is consumed, sending the user to
 * /sso-link; otherwise an explicit redirect wins, and without one a device enrolment code left by
 * an accept-device link is consumed and takes precedence over the dashboard, so the link the user
 * originally followed still completes.
 */
export function resolvePostLoginRedirect(params: URLSearchParams): string {
  return landingAfterSignIn(getSafeRedirect(params));
}

/**
 * Where an already validated redirect lands once the sign-in succeeded. A pending SSO link wins
 * and is consumed, sending the user to confirm it; otherwise it consumes the pending device code.
 * Call it only after the session exists, never while rendering.
 */
export function landingAfterSignIn(redirect: string): string {
  if (sessionStorage.getItem(PENDING_SSO_LINK_KEY)) {
    clearPendingSsoLink();
    return "/sso-link";
  }
  if (redirect === "/dashboard") {
    const pendingCode = consumePendingDeviceCode();
    if (pendingCode) {
      return `/accept-device?code=${encodeURIComponent(pendingCode)}`;
    }
  }
  return redirect;
}
