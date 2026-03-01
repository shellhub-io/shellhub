import { useState, useEffect, useRef } from "react";

export function useCountdown(targetTimestamp: number | null) {
  const [timeLeft, setTimeLeft] = useState<string>("");
  const [isExpired, setIsExpired] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!targetTimestamp) return;

    // Reset expired state when receiving a new timestamp
    setIsExpired(false);
    setTimeLeft("");

    const updateCountdown = () => {
      const now = Date.now();
      const diff = targetTimestamp * 1000 - now; // Convert to ms

      if (diff <= 0) {
        setTimeLeft("0 seconds");
        setIsExpired(true);
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
        }
        return;
      }

      const minutes = Math.floor(diff / 60000);
      const seconds = Math.floor((diff % 60000) / 1000);
      setTimeLeft(
        `${minutes} minute${minutes !== 1 ? "s" : ""} ${seconds} second${seconds !== 1 ? "s" : ""}`
      );
    };

    updateCountdown();
    intervalRef.current = setInterval(updateCountdown, 1000);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [targetTimestamp]);

  return { timeLeft, isExpired };
}
