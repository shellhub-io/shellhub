import { useState } from "react";
import { getSessionRecord } from "@/client";

export function useSessionRecording() {
  const [logs, setLogs] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = async (uid: string): Promise<boolean> => {
    setIsLoading(true);
    setError(null);
    try {
      const { data } = await getSessionRecord({
        path: { uid, seat: 0 },
        parseAs: "text",
        throwOnError: true,
      });
      const recording: unknown = data;
      if (typeof recording !== "string") throw new Error("recording is not text");

      setLogs(recording);
      return true;
    } catch {
      setError("Failed to load recording");
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const clearLogs = () => {
    setLogs(null);
    setError(null);
  };

  return { logs, isLoading, error, fetchLogs, clearLogs };
}
