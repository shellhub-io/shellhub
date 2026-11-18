import { heldByBrowser, useRecordingsStore } from "@/stores/recordingsStore";
import type { Permission } from "@/utils/permission";

/**
 * Whether a session's recording is available anywhere, and the permission each action on one seat's
 * recording needs, seat zero by default, undefined when it needs none. Reading it, to play or
 * download, takes the browser's copy when there is one and the seat is zero, since that copy is
 * seat zero's and the user's own, and needs SessionPlay when the server's copy is what it would
 * read. Deleting reaches the server's copy whenever recorded is set, so it needs SessionRemove
 * then.
 */
export function useRecordingPermissions(
  sessionUid: string,
  recorded: boolean,
  seat = 0,
): { available: boolean; read?: Permission; remove?: Permission } {
  const local = useRecordingsStore(
    (s) => heldByBrowser(sessionUid)(s) !== undefined,
  );
  const localSeat = useRecordingsStore(
    (s) => heldByBrowser(sessionUid, seat)(s) !== undefined,
  );
  return {
    available: local || recorded,
    read: !localSeat && recorded ? "SessionPlay" : undefined,
    remove: recorded ? "SessionRemove" : undefined,
  };
}
