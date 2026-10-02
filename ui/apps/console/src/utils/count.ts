const numberFormat = new Intl.NumberFormat();

/**
 * A count as a person reads it, grouped for their locale (12,345 or 12.345).
 */
export function formatCount(value: number): string {
  return numberFormat.format(value);
}
