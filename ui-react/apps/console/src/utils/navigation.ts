/**
 * Extracts and validates a redirect path from URL search params.
 * Returns the path only if it's a safe, relative path (no protocol-relative URLs).
 * Falls back to the provided default (or "/dashboard").
 */
export function getSafeRedirect(
  params: URLSearchParams,
  fallback = "/dashboard",
): string {
  const raw = params.get("redirect");
  if (raw && raw.startsWith("/") && !raw.startsWith("//") && !raw.startsWith("/\\")) {
    return raw;
  }
  return fallback;
}
