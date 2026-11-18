import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { getSessionRecord, type Session } from "@/client";
import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { sessionTitle } from "@/utils/session";
import {
  castFilename,
  readRecording,
  recordingId,
  saveRecording,
} from "@/utils/recordings";

async function fetchRecording(uid: string, seat: number): Promise<string> {
  const { data } = await getSessionRecord({
    path: { uid, seat },
    parseAs: "text",
    throwOnError: true,
  });
  const recording: unknown = data;
  if (typeof recording !== "string") throw new Error("recording is not text");
  return recording;
}

function readSessionRecording(uid: string, seat: number): Promise<string> {
  const local = heldByBrowser(uid, seat)(useRecordingsStore.getState());
  return local ? readRecording(local) : fetchRecording(uid, seat);
}

const STILL_ACTIVE =
  "The session is still active. Its recording is available when it ends.";

function failureMessage(err: unknown, fallback: string): string {
  return isSdkError(err) && err.status === 409 ? STILL_ACTIVE : fallback;
}

function recordingFilename(session: Session, seat: number): string {
  return castFilename(
    session.device?.name ?? session.device_uid ?? "",
    new Date(session.started_at),
    seat,
  );
}

/**
 * Plays or downloads one seat's recording, seat zero by default, reading the copy the browser holds
 * when there is one, which is seat zero's, and the server's otherwise. play opens it in its own
 * tab, or brings an open one forward without reading it again; download saves it as an asciicast
 * file named after the device and start time, and after the seat when it is not zero. Not a
 * query: a recording is large and only wanted when it is played or downloaded, so caching it with
 * the page would pull it for every listed session. play resolves false, with error set, when the
 * recording could not be read. deleted lists the recordingId of every seat the server
 * answered 404 for, which is how a seat whose recording was deleted is told apart, since the server
 * keeps no per-seat flag. Both refuse a session that is still active, whatever copy exists, because
 * its recording is partial.
 */
export function useSessionRecording() {
  const [isLoading, setIsLoading] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deleted, setDeleted] = useState<string[]>([]);

  const noteDeleted = (err: unknown, id: string) => {
    if (isSdkError(err) && err.status === 404) {
      setDeleted((ids) => (ids.includes(id) ? ids : [...ids, id]));
    }
  };

  const play = async (session: Session, seat = 0): Promise<boolean> => {
    if (session.active) {
      setError(STILL_ACTIVE);
      return false;
    }

    const id = recordingId(session.uid, seat);
    const terminals = useTerminalStore.getState();
    if (terminals.recordings.some((r) => r.id === id)) {
      setError(null);
      terminals.showRecording(id);
      return true;
    }

    setIsLoading(true);
    setError(null);
    try {
      const logs = await readSessionRecording(session.uid, seat);
      useTerminalStore.getState().openRecording({
        id,
        sessionUid: session.uid,
        seat,
        title: sessionTitle(session),
        tenant: session.tenant_id,
        logs,
        filename: recordingFilename(session, seat),
        recorded: session.recorded,
      });
      return true;
    } catch (err: unknown) {
      noteDeleted(err, id);
      setError(failureMessage(err, "Failed to load recording"));
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const download = async (session: Session, seat = 0): Promise<void> => {
    if (session.active) {
      setError(STILL_ACTIVE);
      return;
    }

    setIsDownloading(true);
    setError(null);
    try {
      saveRecording(
        await readSessionRecording(session.uid, seat),
        recordingFilename(session, seat),
      );
    } catch (err: unknown) {
      noteDeleted(err, recordingId(session.uid, seat));
      setError(failureMessage(err, "Failed to download recording"));
    } finally {
      setIsDownloading(false);
    }
  };

  return { isLoading, isDownloading, error, deleted, play, download };
}
