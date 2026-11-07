import { useEffect, useState } from "react";

/**
 * Returns a debounced copy of `value` that updates `delayMs` after the input
 * stops changing.
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
