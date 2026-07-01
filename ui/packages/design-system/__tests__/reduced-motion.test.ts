// @vitest-environment node
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, it, expect } from "vitest";

const baseCss = readFileSync(
  fileURLToPath(new URL("../css/base.css", import.meta.url)),
  "utf8",
);

describe("base.css prefers-reduced-motion reset", () => {
  it("contains a @media (prefers-reduced-motion: reduce) block", () => {
    expect(baseCss).toMatch(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
  });

  it("targets *, *::before, *::after inside the media block", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/\*\s*,\s*\*::before\s*,\s*\*::after/);
  });

  it("contains animation-duration with !important", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/animation-duration\s*:\s*[^;]+!important/);
  });

  it("contains animation-delay with !important", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/animation-delay\s*:\s*[^;]+!important/);
  });

  it("contains animation-iteration-count with !important", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/animation-iteration-count\s*:\s*[^;]+!important/);
  });

  it("contains transition-duration with !important", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/transition-duration\s*:\s*[^;]+!important/);
  });

  it("contains scroll-behavior: auto with !important", () => {
    const mediaIdx = baseCss.search(/@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/);
    expect(mediaIdx).toBeGreaterThan(-1);
    const afterMedia = baseCss.slice(mediaIdx);
    expect(afterMedia).toMatch(/scroll-behavior\s*:\s*auto\s*!important/);
  });
});
