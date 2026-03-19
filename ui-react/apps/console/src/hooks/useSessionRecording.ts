import { useState } from "react";
import apiClient from "../api/client";

export function useSessionRecording() {
  const [logs, setLogs] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = async (uid: string): Promise<boolean> => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<string>(`/api/sessions/${uid}/records/0`);
      setLogs(response.data);
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
