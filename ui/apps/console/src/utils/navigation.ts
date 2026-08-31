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
  return localStorage.getItem(PENDING_DEVICE_CODE_KEY) !== null;
}

/**
 * Reads the pending enrolment code and clears it in the same step, so a code is acted on once.
 * Returns null when there is none.
 */
export function consumePendingDeviceCode(): string | null {
  const code = localStorage.getItem(PENDING_DEVICE_CODE_KEY);
  if (code) clearPendingDeviceCode();
  return code;
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
 * Where to land after a sign-in. An explicit redirect wins; otherwise a device enrolment code
 * left by an accept-device link is consumed and takes precedence over the dashboard, so the
 * link the user originally followed still completes.
 */
export function resolvePostLoginRedirect(params: URLSearchParams): string {
  const redirect = getSafeRedirect(params);
  if (redirect === "/dashboard") {
    const pendingCode = consumePendingDeviceCode();
    if (pendingCode) {
      return `/accept-device?code=${encodeURIComponent(pendingCode)}`;
    }
  }
  return redirect;
}
