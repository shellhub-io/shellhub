import { useEffect, useState } from "react";

/**
 * value once it has stopped changing for ms, so a burst of changes, like a slider being dragged,
 * reaches something expensive only once. The first value is returned at once.
 */
export function useSettledValue<T>(value: T, ms: number): T {
  const [settled, setSettled] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setSettled(value), ms);
    return () => clearTimeout(timer);
  }, [value, ms]);

  return settled;
}
