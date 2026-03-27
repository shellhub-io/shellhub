import { useState, useCallback } from "react";
import { getUserTokenAdmin } from "../client";

export function useLoginAsUser() {
  const [loadingId, setLoadingId] = useState<string | null>(null);
  const [errorId, setErrorId] = useState<string | null>(null);

  const loginAs = useCallback(
    async (userId: string) => {
      if (loadingId) return;
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
        setLoadingId(null);
      }
    },
    [loadingId],
  );

  return { loginAs, loadingId, errorId };
}
