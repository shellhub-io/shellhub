/**
 * Formats a device limit for display. -1 is the sentinel for no limit, not a count.
 */
export function formatMaxDevices(value: number): string {
  return value === -1 ? "Unlimited" : String(value);
}
