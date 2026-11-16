/**
 * The Unix time in seconds at which an account lockout ends, read from the `X-Account-Lockout`
 * header of a 429. A missing header, or one that is zero or not a number, gives null.
 */
export function lockoutEndFrom(headers: Headers | undefined): number | null {
  const epoch = Number(headers?.get("x-account-lockout"));

  return epoch > 0 ? epoch : null;
}

/**
 * The time left in a lockout as a person reads it: whole minutes from a minute up, whole seconds
 * below it, both rounded down, so every lockout message counts the same way.
 */
export function formatTimeLeft(seconds: number): string {
  const [count, unit] = seconds < 60
    ? [Math.floor(seconds), "second"]
    : [Math.floor(seconds / 60), "minute"];

  return `${count} ${unit}${count === 1 ? "" : "s"}`;
}
