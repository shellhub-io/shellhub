import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import type { Action } from "@/utils/permission";

/**
 * Whether a session's recording is available anywhere, and the permission each action on it
 * needs, undefined when it needs none. Reading it, to play or download, takes the browser's copy
 * when there is one, which is the user's own, and needs session:play only when the server's copy is
 * what it would read. Deleting reaches the server's copy whenever recorded is set, so it needs
 * session:removeRecord then.
 */
export function useRecordingPermissions(
  sessionUid: string,
  recorded: boolean,
): { available: boolean; read?: Action; remove?: Action } {
  const local = useRecordingsStore(
    (s) => heldByBrowser(sessionUid)(s) !== undefined,
  );
  return {
    available: local || recorded,
    read: !local && recorded ? "session:play" : undefined,
    remove: recorded ? "session:removeRecord" : undefined,
  };
}
