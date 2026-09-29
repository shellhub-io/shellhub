import { useDeleteSessionRecording } from "@/hooks/useSessionMutations";
import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import { useTerminalStore } from "@/stores/terminalStore";

/**
 * Deletes every copy of a session's recording: the one the browser holds, and the server's when
 * recorded is set. The tab playing it is closed. Rejects when a copy could not be deleted, with the
 * browser's copy already gone if that is the one that went.
 */
export function useRemoveRecording() {
  const deleteServerRecording = useDeleteSessionRecording();
  const removeLocalRecording = useRecordingsStore((s) => s.remove);

  return async (sessionUid: string, recorded: boolean): Promise<void> => {
    const local = heldByBrowser(sessionUid)(useRecordingsStore.getState());
    if (local) await removeLocalRecording(local.id);
    if (recorded) await deleteServerRecording.mutateAsync(sessionUid);
    useTerminalStore.getState().closeRecording(sessionUid);
  };
}
