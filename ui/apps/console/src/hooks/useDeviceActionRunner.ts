import { useCallback } from "react";
import {
  useAcceptDevice,
  useUpdateDeviceStatus,
  useDeleteDevice,
} from "@/client/api";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/**
 * Runs the confirmation dialog's chosen operation against a device, so the dialog does not have
 * to know which mutation each operation maps to.
 */
export function useDeviceActionRunner() {
  const accept = useAcceptDevice();
  const reject = useUpdateDeviceStatus();
  const remove = useDeleteDevice();

  return useCallback(
    async (entity: EntityBase, operation: EntityOperation) => {
      if (operation === "reject") {
        await reject.mutateAsync({
          path: { uid: entity.uid, status: "reject" },
        });
        return;
      }

      const action = operation === "accept" ? accept : remove;
      await action.mutateAsync({ path: { uid: entity.uid } });
    },
    [accept, reject, remove],
  );
}
