const numberFormat = new Intl.NumberFormat();

/**
 * A count as a person reads it, grouped for their locale (12,345 or 12.345).
 */
export function formatCount(value: number): string {
  return numberFormat.format(value);
}

/**
 * pluralize returns a count followed by its unit, with an "s" added to the unit unless the count
 * is exactly 1: "1 minute", "0 seconds". The count is shown as given, so the caller rounds it
 * first, and a unit whose plural is not a trailing "s" needs its own wording.
 */
export function pluralize(count: number, unit: string): string {
  return `${count} ${unit}${count === 1 ? "" : "s"}`;
}
