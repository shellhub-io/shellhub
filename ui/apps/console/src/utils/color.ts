function relativeLuminance(hex: string): number {
  const value = hex.replace("#", "").slice(0, 6);
  const [r, g, b] = [0, 2, 4].map((i) => {
    const channel = parseInt(value.slice(i, i + 2), 16) / 255;
    return channel <= 0.03928
      ? channel / 12.92
      : ((channel + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/**
 * The WCAG contrast ratio between two #rrggbb colours, from 1 (identical) to 21 (black on white).
 * An alpha suffix is ignored.
 */
export function contrastRatio(a: string, b: string): number {
  const [light, dark] = [relativeLuminance(a), relativeLuminance(b)].sort(
    (x, y) => y - x,
  );
  return (light + 0.05) / (dark + 0.05);
}

/**
 * The first of candidates that reaches 3:1 against background, or fallback when none does.
 * Candidates go in order of preference, so a theme keeps its intended hue whenever it is legible.
 */
export function readableOn(
  background: string,
  candidates: string[],
  fallback: string,
): string {
  return (
    candidates.find((c) => contrastRatio(c, background) >= 3) ?? fallback
  );
}
