import { describe, it, expect } from "vitest";
import { contrastRatio, readableOn } from "../color";

describe("contrastRatio", () => {
  it.each([
    ["#000000", "#ffffff", 21],
    ["#ffffff", "#000000", 21],
    ["#777777", "#777777", 1],
  ])("rates %s against %s as %d", (a, b, ratio) => {
    expect(contrastRatio(a, b)).toBeCloseTo(ratio, 1);
  });

  it("ignores an alpha suffix", () => {
    expect(contrastRatio("#00000080", "#ffffff")).toBeCloseTo(21, 1);
  });
});

describe("readableOn", () => {
  const dark = "#1e1e1e";

  it("keeps the first candidate that reads", () => {
    expect(readableOn(dark, ["#4e9a06", "#8ae234"], "#ffffff")).toBe(
      "#4e9a06",
    );
  });

  it("skips a candidate too close to the background", () => {
    expect(readableOn(dark, ["#2a2a2a", "#8ae234"], "#ffffff")).toBe(
      "#8ae234",
    );
  });

  it("falls back when no candidate reads", () => {
    expect(readableOn(dark, ["#222222", "#262626"], "#ffffff")).toBe(
      "#ffffff",
    );
  });
});
