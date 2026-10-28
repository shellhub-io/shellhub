/**
 * Reduces what a user typed or pasted to the bare pairing code: the `code` parameter when the
 * input is the link the agent prints, otherwise the input itself, uppercased and stripped of
 * anything that is not a letter or a digit, as the server normalizes it.
 */
export function normalizePairingCode(input: string): string {
  const fromLink = /[?&]code=([^&#\s]+)/i.exec(input)?.[1];
  return (fromLink ?? input).toUpperCase().replace(/[^0-9A-Z]/g, "");
}

/**
 * Normalizes input as normalizePairingCode does and returns the code when it has the shape the
 * server issues, eight characters from its alphabet, which leaves out the ones easily misread
 * (0, 1, I, L, O, U). Returns null for anything else, an empty or missing input included.
 */
export function parsePairingCode(input: string | null): string | null {
  const code = normalizePairingCode(input ?? "");
  return /^[2-9A-HJKMNP-TV-Z]{8}$/.test(code) ? code : null;
}

/**
 * Shows a bare code the way the agent prints it, `XXXX-XXXX`.
 */
export function formatPairingCode(code: string): string {
  return code.length > 4 ? `${code.slice(0, 4)}-${code.slice(4)}` : code;
}
