import { useState, useEffect } from "react";

import { formatTimeLeft, lockoutEndFrom } from "@/utils/lockout";

interface CountdownState {
  display: string;
  expired: boolean;
  epoch: number | null;
}

function useLockoutCountdown(lockoutEndEpoch: number | null) {
  const [state, setState] = useState<CountdownState>({
    display: "",
    expired: false,
    epoch: null,
  });

  useEffect(() => {
    if (lockoutEndEpoch === null) return;

    const interval = setInterval(() => {
      const diff = lockoutEndEpoch - Date.now() / 1000;
      if (diff <= 0) {
        clearInterval(interval);
        setState({ display: "", expired: true, epoch: lockoutEndEpoch });
      } else {
        setState({
          display: formatTimeLeft(diff),
          expired: false,
          epoch: lockoutEndEpoch,
        });
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [lockoutEndEpoch]);

  if (state.epoch !== lockoutEndEpoch) return { display: "", expired: false };

  return { display: state.display, expired: state.expired };
}

/**
 * Holds a form's account lockout. lockFrom starts it from a 429's headers, ending at the Unix time
 * in seconds that `X-Account-Lockout` carries; a missing, zero or malformed header shows the
 * lockout without a countdown rather than one that already ended. clearLockout ends it before the
 * next try, and lockedOut is true while it runs. countdownDisplay is the time left in whole minutes
 * or seconds, empty until the first tick, and lockoutExpired turns true once it ends; both reset
 * when a new lockout replaces the old one, so a second lockout never reads as already finished.
 */
export function useLockout() {
  const [lockoutEndEpoch, setLockoutEndEpoch] = useState<number | null>(null);
  const { display, expired } = useLockoutCountdown(lockoutEndEpoch);

  return {
    lockFrom: (headers: Headers) => setLockoutEndEpoch(lockoutEndFrom(headers)),
    clearLockout: () => setLockoutEndEpoch(null),
    lockedOut: lockoutEndEpoch !== null && !expired,
    countdownDisplay: display,
    lockoutExpired: expired,
  };
}
