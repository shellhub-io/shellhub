/**
 * Body typeface, as a CSS font stack. Duplicated into the Tailwind preset, which cannot import
 * from here; change both together.
 */
export const FONT_SANS = "IBM Plex Sans, sans-serif";
/**
 * Typeface for code, identifiers and anything a user may need to copy exactly.
 */
export const FONT_MONO = "IBM Plex Mono, monospace";

/**
 * The raw palette, for the few places that need a literal colour — an inline SVG fill, a canvas,
 * a box-shadow Tailwind cannot express. Everywhere else style with the semantic tokens, or the
 * result will not follow the light theme.
 */
export const C = {
  primary: "#667ACC",
  primaryDim: "#667ACC20",
  primaryGlow: "#667ACC40",
  cyan: "#4e9aa3",
  cyanDim: "#4e9aa320",
  yellow: "#bf8c5d",
  yellowDim: "#bf8c5d20",
  green: "#82a568",
  greenDim: "#82a56820",
  red: "#D8737B",
  redDim: "#D8737B20",
  blue: "#56a2e1",
  blueDim: "#56a2e120",
  bg: "#18191B",
  surface: "#1E2127",
  card: "#22252B",
  border: "#2C2F36",
  borderLight: "#383D47",
  text: "#E1E4EA",
  textSec: "#8B8F99",
  textMuted: "#81879C",
};
