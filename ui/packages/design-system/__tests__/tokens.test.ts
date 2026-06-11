import { describe, it, expect } from "vitest";
import rawPreset from "../tailwind.preset.js";
import { C } from "../constants";

interface PresetColors {
  primary: { DEFAULT: string; [key: string]: string };
  background: string;
  surface: string;
  card: string;
  border: string;
  "border-light": string;
  "text-primary": string;
  "text-secondary": string;
  "text-muted": string;
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
 * Base-color mappings: [C key, preset path description, preset value]
 *
 * C values are hand-maintained literals — the point of this test is to
 * catch future hand-edits to constants.ts that drift from the preset.
 * Do NOT derive C values from the preset here.
 */
const baseColorMappings: Array<[keyof typeof C, string, string]> = [
  ["primary", "primary.DEFAULT", colors.primary.DEFAULT],
  ["cyan", "accent.cyan", colors.accent.cyan],
  ["yellow", "accent.yellow", colors.accent.yellow],
  ["green", "accent.green", colors.accent.green],
  ["red", "accent.red", colors.accent.red],
  ["blue", "accent.blue", colors.accent.blue],
  ["bg", "background", colors.background],
  ["surface", "surface", colors.surface],
  ["card", "card", colors.card],
  ["border", "border", colors.border],
  ["borderLight", "border-light", colors["border-light"]],
  ["text", "text-primary", colors["text-primary"]],
  ["textSec", "text-secondary", colors["text-secondary"]],
  ["textMuted", "text-muted", colors["text-muted"]],
];

describe("design-system token parity: C constants match tailwind.preset.js", () => {
  describe("base colors", () => {
    it.each(baseColorMappings)(
      "C.%s matches preset %s",
      (cKey, _presetPath, presetValue) => {
        expect(C[cKey]).toBe(presetValue);
      },
    );
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
