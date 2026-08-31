import { useState, useCallback, useRef } from "react";
import { getUserTokenAdmin } from "../client";

/**
 * Signs an admin in as another user. Tracks the id in flight so a row can show its own spinner,
 * and guards against a second click starting a second impersonation.
 */
export function useLoginAsUser() {
  const [loadingId, setLoadingId] = useState<string | null>(null);
  const [errorId, setErrorId] = useState<string | null>(null);
  const loadingRef = useRef(false);

  const loginAs = useCallback(async (userId: string) => {
    if (loadingRef.current) return;
    loadingRef.current = true;
    setLoadingId(userId);
    setErrorId(null);
    try {
      const { data } = await getUserTokenAdmin({
        path: { id: userId },
        throwOnError: true,
      });
      if (data?.token) {
        window.open(
          `/login?token=${encodeURIComponent(data.token)}`,
          "_blank",
          "noopener,noreferrer",
        );
      }
    } catch {
      setErrorId(userId);
    } finally {
      loadingRef.current = false;
      setLoadingId(null);
    }
  }, []);

  return { loginAs, loadingId, errorId };
}
