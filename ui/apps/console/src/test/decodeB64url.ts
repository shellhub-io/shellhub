import { Buffer } from "buffer";

/**
 * Decode a base64url string (RFC 4648 §5) produced by {@link toBase64Json}
 * back to a parsed JS value.
 *
 * The standard `atob()` and `Buffer.from(s, "base64")` only handle the
 * standard alphabet (+/). Base64url uses - and _ in their place, so we
 * normalise the alphabet before decoding.
 */
export function decodeB64url(b64url: string): unknown {
  const standard = b64url.replace(/-/g, "+").replace(/_/g, "/");
  return JSON.parse(Buffer.from(standard, "base64").toString("utf-8"));
}
