import { Buffer } from "buffer";

// `btoa(JSON.stringify(value))` throws InvalidCharacterError for any character
// above U+00FF (e.g. tag names, hostnames, usernames in non-Latin scripts),
// so the request never leaves the browser. UTF-8-encode first, then convert
// standard base64 to unpadded base64url so the result is safe in URL
// query-strings and HTTP headers without percent-encoding.
// Note: feross/buffer v5.7.1 (pinned in the lockfile) does not support
// `.toString('base64url')`, so we post-process the standard base64 string.
export function toBase64Json(value: unknown): string {
  return Buffer.from(JSON.stringify(value), "utf-8")
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}
