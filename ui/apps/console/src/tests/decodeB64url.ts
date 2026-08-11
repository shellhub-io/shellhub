import { Buffer } from "buffer";

export function decodeB64url(b64url: string): unknown {
  const standard = b64url.replace(/-/g, "+").replace(/_/g, "/");
  return JSON.parse(Buffer.from(standard, "base64").toString("utf-8"));
}
