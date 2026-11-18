import { useMutation } from "@tanstack/react-query";
import { deleteSessionRecord } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import { useTerminalStore } from "@/stores/terminalStore";

/**
 * Deletes every copy of a session's recording: the one the browser holds, and the server's seat
 * zero when recorded is set, the only seat the UI records. The tab playing it is closed. Rejects when a copy could not be deleted, with the
 * browser's copy already gone if that is the one that went.
 */
export function useRemoveRecording() {
  const invalidateSessions = useInvalidateByIds("getSessions", "getSession");
  const deleteServerRecording = useMutation({
    mutationFn: async (uid: string) => {
      await deleteSessionRecord({ path: { uid, seat: 0 }, throwOnError: true });
    },
    onSuccess: invalidateSessions,
  });
  const removeLocalRecording = useRecordingsStore((s) => s.remove);

  return async (sessionUid: string, recorded: boolean): Promise<void> => {
    const local = heldByBrowser(sessionUid)(useRecordingsStore.getState());
    if (local) await removeLocalRecording(local.id);
    if (recorded) await deleteServerRecording.mutateAsync(sessionUid);
    useTerminalStore.getState().closeRecording(sessionUid);
  };
}
