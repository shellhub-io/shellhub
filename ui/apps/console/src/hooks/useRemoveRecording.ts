import { useMutation } from "@tanstack/react-query";
import { deleteSessionRecord, getSession } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { recordingId } from "@/utils/recordings";

async function stillRecorded(uid: string): Promise<boolean> {
  const { data } = await getSession({ path: { uid }, throwOnError: true });
  return data.recorded;
}

/**
 * Deletes every copy of the given seats' recordings, seat zero alone when seats is missing or
 * empty: the one the browser holds, which is seat zero's, and the server's when recorded is set.
 * The server's copies go one seat at a time, in order. The first seat trusts recorded; before each
 * later seat the session is read again, and the rest are skipped once it reports no recording left,
 * since the server refuses a delete on a session that is not recorded. Each seat's tab closes as
 * its server copy is deleted, and every given seat's tab once the loop ends.
 *
 * Rejects when the browser's copy could not be deleted, when a server copy could not be deleted
 * (400 when the session no longer holds a recording, as when recorded was stale or another delete
 * took the last seat first, 404 when the session does not exist, 409 while it is still active, 503
 * when the object storage is unavailable, or any other request failure), or when the session could
 * not be read between two deletes. The copies before the failure are already gone. A seat whose
 * server copy was deleted has its tab closed, and a seat whose server copy remains keeps its tab
 * open, even when the browser's copy of seat zero is already gone. When the browser's copy could
 * not be deleted, every tab stays open.
 */
export function useRemoveRecording() {
  const invalidateSessions = useInvalidateByIds("getSessions", "getSession");
  const deleteServerRecording = useMutation({
    mutationFn: async ({ uid, seat }: { uid: string; seat: number }) => {
      await deleteSessionRecord({ path: { uid, seat }, throwOnError: true });
    },
    onSuccess: invalidateSessions,
  });
  const removeLocalRecording = useRecordingsStore((s) => s.remove);
  const closeTab = (sessionUid: string, seat: number) =>
    useTerminalStore.getState().closeRecording(recordingId(sessionUid, seat));

  return async (
    sessionUid: string,
    recorded: boolean,
    requestedSeats?: number[],
  ): Promise<void> => {
    const seats = requestedSeats?.length ? requestedSeats : [0];
    const state = useRecordingsStore.getState();
    const local = seats
      .map((seat) => heldByBrowser(sessionUid, seat)(state))
      .find((copy) => copy !== undefined);
    if (local) await removeLocalRecording(local.id);

    if (recorded) {
      for (const [i, seat] of seats.entries()) {
        if (i > 0 && !(await stillRecorded(sessionUid))) break;
        await deleteServerRecording.mutateAsync({ uid: sessionUid, seat });
        closeTab(sessionUid, seat);
      }
    }

    for (const seat of seats) closeTab(sessionUid, seat);
  };
}
