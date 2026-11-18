import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import type { Permission } from "@/utils/permission";

/**
 * Whether a session's recording is available anywhere, and the permission each action on it
 * needs, undefined when it needs none. Reading it, to play or download, takes the browser's copy
 * when there is one, which is the user's own, and needs SessionPlay only when the server's copy is
 * what it would read. Deleting reaches the server's copy whenever recorded is set, so it needs
 * SessionRemove then.
 */
export function useRecordingPermissions(
  sessionUid: string,
  recorded: boolean,
): { available: boolean; read?: Permission; remove?: Permission } {
  const local = useRecordingsStore(
    (s) => heldByBrowser(sessionUid)(s) !== undefined,
  );
  return {
    available: local || recorded,
    read: !local && recorded ? "SessionPlay" : undefined,
    remove: recorded ? "SessionRemove" : undefined,
  };
}
