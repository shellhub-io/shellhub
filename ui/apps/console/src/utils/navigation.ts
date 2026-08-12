export const PENDING_DEVICE_CODE_KEY = "shellhub:pending-device-code";

export function setPendingDeviceCode(code: string): void {
  localStorage.setItem(PENDING_DEVICE_CODE_KEY, code);
}

export function clearPendingDeviceCode(): void {
  localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
}

export function hasPendingDeviceCode(): boolean {
  return localStorage.getItem(PENDING_DEVICE_CODE_KEY) !== null;
}

export function consumePendingDeviceCode(): string | null {
  const code = localStorage.getItem(PENDING_DEVICE_CODE_KEY);
  if (code) clearPendingDeviceCode();
  return code;
}

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
