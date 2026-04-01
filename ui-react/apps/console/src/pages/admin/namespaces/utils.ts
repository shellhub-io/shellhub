export function formatMaxDevices(value: number): string {
  return value === -1 ? "Unlimited" : String(value);
}
