import { useState } from "react";
import { getSessionRecord, type Session } from "@/client";
import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { sessionTitle } from "@/utils/session";
import { castFilename, readRecording, saveRecording } from "@/utils/recordings";

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

function readSessionRecording(uid: string): Promise<string> {
  const local = heldByBrowser(uid)(useRecordingsStore.getState());
  return local ? readRecording(local) : fetchRecording(uid);
}

function recordingFilename(session: Session): string {
  return castFilename(
    session.device?.name ?? session.device_uid ?? "",
    new Date(session.started_at),
  );
}

/**
 * Plays or downloads a session's recording, reading the copy the browser holds when there is one
 * and the server's otherwise. play opens it in its own tab, or brings an open one forward without
 * reading it again; download saves it as an asciicast file. Not a query: a recording is large and
 * only wanted when it is played or downloaded, so caching it with the page would pull it for every
 * listed session. play resolves false, with error set, when the recording could not be read.
 */
export function useSessionRecording() {
  const [isLoading, setIsLoading] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const play = async (session: Session): Promise<boolean> => {
    const terminals = useTerminalStore.getState();
    if (terminals.recordings.some((r) => r.id === session.uid)) {
      setError(null);
      terminals.showRecording(session.uid);
      return true;
    }

    setIsLoading(true);
    setError(null);
    try {
      const logs = await readSessionRecording(session.uid);
      useTerminalStore.getState().openRecording({
        id: session.uid,
        title: sessionTitle(session),
        tenant: session.tenant_id,
        logs,
        filename: recordingFilename(session),
        recorded: session.recorded,
      });
      return true;
    } catch {
      setError("Failed to load recording");
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const download = async (session: Session): Promise<void> => {
    setIsDownloading(true);
    setError(null);
    try {
      saveRecording(
        await readSessionRecording(session.uid),
        recordingFilename(session),
      );
    } catch {
      setError("Failed to download recording");
    } finally {
      setIsDownloading(false);
    }
  };

  return { isLoading, isDownloading, error, play, download };
}
