import { useState, useEffect } from "react";

interface CountdownState {
  display: string;
  expired: boolean;
  epoch: number | null;
}

/**
 * Reads the end of an account lockout from a 429 response's `X-Account-Lockout` header, as the Unix
 * timestamp in seconds that {@link useLockoutCountdown} takes. A missing, zero or malformed header
 * gives `null`, so the form shows the lockout without a countdown rather than one that already ended.
 */
export function lockoutEndFrom(headers: Headers): number | null {
  const epoch = Number(headers.get("x-account-lockout"));

  return epoch > 0 ? epoch : null;
}

/**
 * Counts down to the end of an account lockout, given as the Unix timestamp in seconds that the
 * `X-Account-Lockout` header carries. `display` is the time left in whole minutes or seconds, empty
 * until the first tick. `expired` turns true once the lockout ends, and both reset when a new
 * lockout replaces the old one, so a second lockout never reads as already finished.
 */
export function useLockoutCountdown(lockoutEndEpoch: number | null) {
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
      } else if (diff < 60) {
        const s = Math.floor(diff);
        setState({
          display: `${s} ${s === 1 ? "second" : "seconds"}`,
          expired: false,
          epoch: lockoutEndEpoch,
        });
      } else {
        const m = Math.floor(diff / 60);
        setState({
          display: `${m} ${m === 1 ? "minute" : "minutes"}`,
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
