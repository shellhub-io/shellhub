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
      // The route serves an asciicast file, not the JSON array the spec declares, so the body
      // is read as text and the generated response type does not describe what arrives. Guard
      // rather than cast, so a spec fix that makes the type honest surfaces here instead of
      // handing the player a value it cannot render.
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
