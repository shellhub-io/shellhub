import { useCallback, useEffect, useRef, useState } from "react";
import { getStatusDevices, type GetStatusDevicesResponse } from "@/client";

interface UseDevicePollingOptions {
  /** Called after each successful poll. Return `true` to stop polling. */
  onPoll: (stats: GetStatusDevicesResponse) => boolean;
  /** Initial delay between polls in ms. Default: 2000 */
  initialInterval?: number;
  /** Maximum delay between polls in ms (backoff cap). Default: 10000 */
  maxInterval?: number;
  /** Multiplier applied after each poll. Default: 1.5 */
  backoffFactor?: number;
}

interface UseDevicePollingReturn {
  isPolling: boolean;
  start: () => void;
  stop: () => void;
}

const noop = (_delay: number) => {};

/**
 * Polls GET /api/stats on a setTimeout chain with exponential backoff.
 *
 * Unlike setInterval, this approach waits for the response before scheduling
 * the next poll, preventing request pile-up when the server is slow.
 * Cleanup on unmount is guaranteed via useEffect return.
 */
export function useDevicePolling({
  onPoll,
  initialInterval = 2000,
  maxInterval = 10000,
  backoffFactor = 1.5,
}: UseDevicePollingOptions): UseDevicePollingReturn {
  const [isPolling, setIsPolling] = useState(false);

  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const activeRef = useRef(false);

  // Keep onPoll in a ref to avoid stale closures in the recursive loop
  const onPollRef = useRef(onPoll);
  useEffect(() => {
    onPollRef.current = onPoll;
  });

  // scheduleRef holds the recursive scheduler so it can call itself without
  // triggering the "accessed before declaration" lint error on useCallback.
  const scheduleRef = useRef<(delay: number) => void>(noop);

  const stop = useCallback(() => {
    activeRef.current = false;
    setIsPolling(false);
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  // Redefine the scheduler whenever backoff params or stop change.
  // Storing in a ref means the recursive call inside setTimeout always
  // picks up the latest version without useCallback circular dependency.
  useEffect(() => {
    scheduleRef.current = (delay: number) => {
      timeoutRef.current = setTimeout(() => {
        void (async () => {
          if (!activeRef.current) return;

          try {
            const { data: stats } = await getStatusDevices({ throwOnError: true });
            if (!activeRef.current) return;

            const shouldStop = onPollRef.current(stats);
            if (shouldStop) {
              stop();
              return;
            }
          } catch {
            // On error, keep polling but back off
          }

          if (!activeRef.current) return;

          const next = Math.min(delay * backoffFactor, maxInterval);
          scheduleRef.current(next);
        })();
      }, delay);
    };
  }, [backoffFactor, maxInterval, stop]);

  const start = useCallback(() => {
    if (activeRef.current) return;
    activeRef.current = true;
    setIsPolling(true);
    scheduleRef.current(initialInterval);
  }, [initialInterval]);

  // Guarantee cleanup when the component using this hook unmounts
  useEffect(() => {
    return () => {
      activeRef.current = false;
      if (timeoutRef.current !== null) clearTimeout(timeoutRef.current);
    };
  }, []);

  return { isPolling, start, stop };
}
