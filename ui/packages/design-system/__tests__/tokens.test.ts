// @vitest-environment node
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, it, expect } from "vitest";
import rawPreset from "../tailwind.preset.js";
import { C } from "../constants";

interface PresetColors {
  primary: { DEFAULT: string; [key: string]: string };
  accent: {
    green: string;
    red: string;
    yellow: string;
    blue: string;
    cyan: string;
  };
}

const preset = rawPreset as unknown as {
  theme: { extend: { colors: PresetColors } };
};

const colors = preset.theme.extend.colors;

/**
 * Semantic tokens (background, surface, text, ...) no longer hold literal
 * colors in the preset — they resolve through `rgb(var(--c-*) / <alpha>)` so
 * the UI can flip between dark and light (see css/base.css). Their canonical
 * dark values now live in the `:root` block of base.css as space-separated RGB
 * channels. Parse those and convert to hex so we can still check that the
 * hand-maintained C constants (used for SVG fills) don't drift from them.
 */
const baseCss = readFileSync(
  fileURLToPath(new URL("../css/base.css", import.meta.url)),
  "utf8",
);

const rootBlock = baseCss.match(/:root\s*\{([\s\S]*?)\}/)?.[1] ?? "";

function rgbChannelsToHex(channels: string): string {
  const hex = channels
    .trim()
    .split(/\s+/)
    .map((n) => Number(n).toString(16).padStart(2, "0"))
    .join("");
  return `#${hex}`.toUpperCase();
}

function darkToken(name: string): string {
  const match = rootBlock.match(new RegExp(`--c-${name}:\\s*([\\d\\s]+);`));
  if (!match) throw new Error(`--c-${name} not found in base.css :root`);
  return rgbChannelsToHex(match[1]);
}

/** Preset-literal base colors: [C key, preset path description, preset value] */
const presetColorMappings: Array<[keyof typeof C, string, string]> = [
  ["primary", "primary.DEFAULT", colors.primary.DEFAULT],
  ["cyan", "accent.cyan", colors.accent.cyan],
  ["yellow", "accent.yellow", colors.accent.yellow],
  ["green", "accent.green", colors.accent.green],
  ["red", "accent.red", colors.accent.red],
  ["blue", "accent.blue", colors.accent.blue],
];

/** CSS-var-driven base colors: [C key, css var name] (dark values from :root) */
const cssVarColorMappings: Array<[keyof typeof C, string]> = [
  ["bg", "background"],
  ["surface", "surface"],
  ["card", "card"],
  ["border", "border"],
  ["borderLight", "border-light"],
  ["text", "text-primary"],
  ["textSec", "text-secondary"],
  ["textMuted", "text-muted"],
];

describe("design-system token parity: C constants match their source of truth", () => {
  describe("preset-literal base colors", () => {
    it.each(presetColorMappings)(
      "C.%s matches preset %s",
      (cKey, _presetPath, presetValue) => {
        expect(C[cKey]).toBe(presetValue);
      },
    );
  });

  describe("css-var-driven base colors match base.css :root (dark)", () => {
    it.each(cssVarColorMappings)("C.%s matches --c-%s", (cKey, varName) => {
      expect(C[cKey].toUpperCase()).toBe(darkToken(varName));
    });
  });

  describe("alpha variants follow the *Dim=base+'20' / primaryGlow=base+'40' convention", () => {
    it("primaryDim = primary + '20'", () => {
      expect(C.primaryDim).toBe(C.primary + "20");
    });

    it("primaryGlow = primary + '40'", () => {
      expect(C.primaryGlow).toBe(C.primary + "40");
    });

    const dimVariants: Array<[keyof typeof C, keyof typeof C]> = [
      ["cyanDim", "cyan"],
      ["yellowDim", "yellow"],
      ["greenDim", "green"],
      ["redDim", "red"],
      ["blueDim", "blue"],
    ];

    it.each(dimVariants)("C.%s = C.%s + '20'", (dimKey, baseKey) => {
      expect(C[dimKey]).toBe(C[baseKey] + "20");
    });
  });
});
