import { useEffect, useState } from "react";

/**
 * Returns a debounced copy of `value`. The first value is returned at once;
 * later values arrive once `value` has stayed the same for `delayMs`. A change
 * to `value` or `delayMs` cancels the pending update and restarts the timer.
 *
 * Objects and arrays must keep a stable reference between changes. A fresh
 * reference on every render restarts the timer.
 */
export function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}
