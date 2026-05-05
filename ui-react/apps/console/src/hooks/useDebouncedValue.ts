import { useEffect, useState } from "react";

/**
 * Returns a debounced copy of `value` that updates `delayMs` after the input
 * stops changing.
 *
 * Use only with **primitive** values (string, number, boolean). Object or array
 * inputs reset the timer on every render that produces a fresh reference,
 * defeating the debounce. For non-primitives, debounce a stable derived key
 * instead (e.g. `JSON.stringify`).
 */
export function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}
