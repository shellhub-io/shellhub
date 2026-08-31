import { Buffer } from "buffer";

/**
 * Decodes the base64url a request carries its filter in, so a test can assert on the filter
 * rather than on an opaque string.
 */
export function decodeB64url(b64url: string): unknown {
  const standard = b64url.replace(/-/g, "+").replace(/_/g, "/");
  return JSON.parse(Buffer.from(standard, "base64").toString("utf-8"));
}
