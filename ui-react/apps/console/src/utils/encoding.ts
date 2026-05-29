import { Buffer } from "buffer";

// `btoa(JSON.stringify(value))` throws InvalidCharacterError for any character
// above U+00FF (e.g. tag names, hostnames, usernames in non-Latin scripts),
// so the request never leaves the browser. UTF-8-encode first, then base64.
export function toBase64Json(value: unknown): string {
  return Buffer.from(JSON.stringify(value), "utf-8").toString("base64");
}
