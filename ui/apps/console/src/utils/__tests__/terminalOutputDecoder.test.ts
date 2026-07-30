import { describe, expect, it } from "vitest";
import { createOutputDecoder } from "@/utils/terminalOutputDecoder";

const bytes = (...values: number[]) => new Uint8Array(values);

describe("createOutputDecoder", () => {
  it("defaults to UTF-8", () => {
    const decoder = createOutputDecoder();

    expect(decoder.decode(bytes(0xe4, 0xb8, 0xad))).toBe("中");
  });

  it("holds a multi-byte character split across chunks until it completes", () => {
    const decoder = createOutputDecoder("utf-8");

    expect(decoder.decode(bytes(0xe4, 0xb8))).toBe("");
    expect(decoder.decode(bytes(0xad))).toBe("中");
  });

  it("decodes a character split across chunks for non-UTF-8 encodings too", () => {
    const decoder = createOutputDecoder("gbk");

    expect(decoder.decode(bytes(0xd6))).toBe("");
    expect(decoder.decode(bytes(0xd0))).toBe("中");
  });

  it("decodes GBK bytes that UTF-8 would destroy", () => {
    const decoder = createOutputDecoder("gbk");

    expect(decoder.decode(bytes(0xd6, 0xd0, 0xce, 0xc4, 0xc4, 0xbf, 0xc2, 0xbc))).toBe("中文目录");
  });

  it.each([
    ["an unknown encoding name", "not-a-real-encoding"],
    ["an empty encoding name", ""],
  ])("falls back to UTF-8 for %s", (_description, encoding) => {
    const decoder = createOutputDecoder(encoding);

    expect(decoder.decode(bytes(0xe4, 0xb8, 0xad))).toBe("中");
  });

  it("yields the replacement character for invalid bytes rather than throwing", () => {
    const decoder = createOutputDecoder("utf-8");

    expect(decoder.decode(bytes(0x41, 0xff, 0x42))).toBe("A�B");
  });

  it("keeps decoding across many chunks", () => {
    const decoder = createOutputDecoder("utf-8");
    const chunks = [bytes(0x68, 0x69, 0xe6), bytes(0x97, 0xa5), bytes(0x21)];

    expect(chunks.map((chunk) => decoder.decode(chunk)).join("")).toBe("hi日!");
  });
});
