import { describe, it, expect } from "vitest";
import { toBase64Json } from "@/utils/encoding";
import { decodeB64url } from "@/test/decodeB64url";

describe("toBase64Json", () => {
  it("emits unpadded base64url (no +, /, or = characters)", () => {
    // The output must be safe to embed in a URL query-string or HTTP header
    // without percent-encoding. Standard base64 uses + and / which are
    // URL-unsafe, and adds = padding which is also forbidden in some contexts.
    // Covers ASCII (padding), Unicode (/ replacement), and mixed payloads.
    const inputs = [
      // ASCII payload — v1 ends with "=" in standard base64
      [{ type: "property", params: { name: "name", operator: "contains", value: "qa-edge" } }],
      // Unicode payload — v2 contains "/" in standard base64
      [{ type: "property", params: { name: "tags.name", operator: "contains", value: "日本語タグ" } }],
      // Emoji + Arabic — also contains "/" in standard base64
      { tag: "日本語タグ", emoji: "🚀", arabic: "سلام" },
    ];
    for (const value of inputs) {
      const encoded = toBase64Json(value);
      expect(encoded, `output for ${JSON.stringify(value).slice(0, 40)}`).not.toMatch(/[+/=]/);
    }
  });

  it("round-trips ASCII payload through unpadded base64url → UTF-8 → JSON", () => {
    // The BE decoder reads the raw bytes, so the decoded JSON must match.
    // After switching to base64url we no longer compare to btoa() — the two
    // alphabets differ. Verify correctness by decoding instead.
    const value = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: "qa-edge" },
      },
    ];
    const decoded = decodeB64url(toBase64Json(value)) as typeof value;
    expect(decoded).toEqual(value);
  });

  it("does not throw on non-Latin-1 characters", () => {
    const value = [
      {
        type: "property",
        params: {
          name: "tags.name",
          operator: "contains",
          value: "日本語タグ",
        },
      },
    ];
    expect(() => toBase64Json(value)).not.toThrow();
  });

  it("round-trips Unicode through base64url → UTF-8 → JSON", () => {
    const value = { tag: "日本語タグ", emoji: "🚀", arabic: "سلام" };
    const decoded = decodeB64url(toBase64Json(value)) as typeof value;
    expect(decoded).toEqual(value);
  });

  it("maps std base64 '+' to '-' and strips '=' padding ('>>>' payload)", () => {
    // '>>>' encodes to "Ij4+PiI=" in standard base64 (contains '+' and '=').
    // base64url must replace '+' with '-' and strip trailing '='.
    const json = JSON.stringify(">>>");
    const expected = btoa(json).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
    expect(toBase64Json(">>>")).toBe(expected);
    expect(toBase64Json(">>>")).not.toMatch(/[+/=]/);
  });

  it("produces stable output for the same input (used as a cache key)", () => {
    const value = [
      { type: "property", params: { name: "x", operator: "eq", value: 1 } },
    ];
    expect(toBase64Json(value)).toBe(toBase64Json(value));
  });
});
