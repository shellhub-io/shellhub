/**
 * Whether a URL is safe to put in an href. Only http, https and mailto pass — this exists to
 * keep javascript: out of a link built from data the user supplied.
 */
export function isAllowedUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    return ["http:", "https:", "mailto:"].includes(parsed.protocol);
  } catch {
    return false;
  }
}
