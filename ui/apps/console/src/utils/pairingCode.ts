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
 * Shows a bare code the way the agent prints it, `XXXX-XXXX`.
 */
export function formatPairingCode(code: string): string {
  return code.length > 4 ? `${code.slice(0, 4)}-${code.slice(4)}` : code;
}
