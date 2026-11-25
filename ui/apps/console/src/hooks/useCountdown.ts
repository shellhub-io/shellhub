import { useState, useEffect } from "react";

/**
 * useCountdown returns the seconds left until endEpoch, a Unix time in seconds, with a fractional
 * part the caller rounds as it shows them. It updates each second and stops at 0. It is null while
 * endEpoch is null or not a finite number, and on the render that receives a new endEpoch, until
 * the count for it starts, so an end never shows a count from before it was set.
 */
export function useCountdown(endEpoch: number | null): number | null {
  const [latest, setLatest] = useState<{ epoch: number; left: number } | null>(null);

  useEffect(() => {
    if (endEpoch === null || !Number.isFinite(endEpoch)) return;

    const update = () => {
      const left = Math.max(0, endEpoch - Date.now() / 1000);
      setLatest({ epoch: endEpoch, left });
      return left;
    };

    const interval =
      update() === 0
        ? undefined
        : setInterval(() => {
            if (update() === 0) clearInterval(interval);
          }, 1000);

    return () => {
      clearInterval(interval);
      setLatest(null);
    };
  }, [endEpoch]);

  return latest?.epoch === endEpoch ? latest.left : null;
}
