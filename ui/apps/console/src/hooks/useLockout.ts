import { useState } from "react";

import { useCountdown } from "@/hooks/useCountdown";
import { formatTimeLeft, lockoutEndFrom } from "@/utils/lockout";

/**
 * Holds a form's account lockout. lockFrom starts it from a 429's headers, ending at the Unix time
 * in seconds that `X-Account-Lockout` carries; a missing, zero or malformed header shows the
 * lockout without a countdown rather than one that already ended. clearLockout ends it before the
 * next try, and lockedOut is true while it runs. countdownDisplay is the time left in whole minutes
 * or seconds, "0 seconds" in the last second, and empty with no lockout, before the count starts
 * and once it ends. lockoutExpired turns true once it ends. Both reset when a new lockout replaces
 * the old one, or when the same lockout end comes back after clearLockout, so a second lockout
 * never reads as already finished.
 */
export function useLockout() {
  const [lockoutEndEpoch, setLockoutEndEpoch] = useState<number | null>(null);
  const secondsLeft = useCountdown(lockoutEndEpoch);
  const expired = secondsLeft === 0;

  return {
    lockFrom: (headers: Headers) => setLockoutEndEpoch(lockoutEndFrom(headers)),
    clearLockout: () => setLockoutEndEpoch(null),
    lockedOut: lockoutEndEpoch !== null && !expired,
    countdownDisplay: secondsLeft ? formatTimeLeft(secondsLeft) : "",
    lockoutExpired: expired,
  };
}
