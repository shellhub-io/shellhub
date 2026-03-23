/**
 * UUID v4 generator with fallback for insecure contexts (plain HTTP).
 *
 * `crypto.randomUUID()` is secure-context-only and throws over plain HTTP.
 * The fallback builds a compliant v4 UUID using `crypto.getRandomValues()`,
 * which works in any context.
 */
export function generateRandomUUID(): string {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();

  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, (c) => {
    const n = Number(c);
    return (
      n ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (n / 4)))
    ).toString(16);
  });
}
