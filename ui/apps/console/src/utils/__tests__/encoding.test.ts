import { describe, it, expect } from "vitest";
import { toBase64Json } from "@/utils/encoding";

describe("toBase64Json", () => {
  it("matches btoa(JSON.stringify(...)) for ASCII payloads", () => {
    // Backward-compat guard: the BE decoder only sees the bytes, so an ASCII
    // payload must produce the exact same string the old `btoa` path produced.
    const value = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: "qa-edge" },
      },
    ];
    expect(toBase64Json(value)).toBe(btoa(JSON.stringify(value)));
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

  it("round-trips Unicode through base64 → UTF-8 → JSON", () => {
    const value = { tag: "日本語タグ", emoji: "🚀", arabic: "سلام" };
    const encoded = toBase64Json(value);
    const decoded = JSON.parse(
      Buffer.from(encoded, "base64").toString("utf-8"),
    ) as typeof value;
    expect(decoded).toEqual(value);
  });

  it("produces stable output for the same input (used as a cache key)", () => {
    const value = [
      { type: "property", params: { name: "x", operator: "eq", value: 1 } },
    ];
    expect(toBase64Json(value)).toBe(toBase64Json(value));
  });
});
