import { describe, it, expect } from "vitest";
import { toBase64Json } from "@/utils/encoding";
import { decodeB64url } from "@/tests/decodeB64url";

describe("toBase64Json", () => {
  it("emits unpadded base64url (no +, /, or = characters)", () => {
    const inputs = [
      [
        {
          type: "property",
          params: { name: "name", operator: "contains", value: "qa-edge" },
        },
      ],
      [
        {
          type: "property",
          params: {
            name: "tags.name",
            operator: "contains",
            value: "日本語タグ",
          },
        },
      ],
      { tag: "日本語タグ", emoji: "🚀", arabic: "سلام" },
    ];
    for (const value of inputs) {
      const encoded = toBase64Json(value);
      expect(
        encoded,
        `output for ${JSON.stringify(value).slice(0, 40)}`,
      ).not.toMatch(/[+/=]/);
    }
  });

  it("round-trips ASCII payload through unpadded base64url → UTF-8 → JSON", () => {
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
    const json = JSON.stringify(">>>");
    const expected = btoa(json)
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");
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
