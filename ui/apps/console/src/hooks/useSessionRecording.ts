import { useState } from "react";
import { getSessionRecord, type Session } from "@/client";
import { useTerminalStore } from "@/stores/terminalStore";
import { sessionTitle } from "@/utils/session";

async function fetchRecording(uid: string): Promise<string> {
  const { data } = await getSessionRecord({
    path: { uid, seat: 0 },
    parseAs: "text",
    throwOnError: true,
  });
  const recording: unknown = data;
  if (typeof recording !== "string") throw new Error("recording is not text");
  return recording;
}

/**
 * Plays a session's recording in its own tab. A recording already open is brought forward
 * without fetching it again; otherwise it is read with readLocal when the browser holds a copy,
 * or fetched from the server. Not a query: a recording is large and only wanted when it is
 * played, so caching it with the page would pull it for every listed session. play resolves
 * false, with error set, when the recording could not be read.
 */
export function useSessionRecording() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const play = async (
    session: Session,
    readLocal?: () => Promise<string>,
  ): Promise<boolean> => {
    const terminals = useTerminalStore.getState();
    if (terminals.recordings.some((r) => r.id === session.uid)) {
      setError(null);
      terminals.showRecording(session.uid);
      return true;
    }

    setIsLoading(true);
    setError(null);
    try {
      const logs = readLocal
        ? await readLocal()
        : await fetchRecording(session.uid);
      useTerminalStore.getState().openRecording({
        id: session.uid,
        title: sessionTitle(session),
        tenant: session.tenant_id,
        logs,
      });
      return true;
    } catch {
      setError("Failed to load recording");
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  return { isLoading, error, play };
}
