import { useCallback } from "react";
import {
  useAcceptDevice,
  useUpdateDeviceStatus,
  useDeleteDevice,
} from "@/client/api";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/** Maps action-dialog operations to device mutations. */
export function useDeviceActionRunner() {
  const accept = useAcceptDevice();
  const reject = useUpdateDeviceStatus();
  const remove = useDeleteDevice();

  return useCallback(
    async (entity: EntityBase, operation: EntityOperation) => {
      if (operation === "reject") {
        await reject.mutateAsync({ uid: entity.uid, status: "reject" });
        return;
      }

      const action = operation === "accept" ? accept : remove;
      await action.mutateAsync({ uid: entity.uid });
    },
    [accept, reject, remove],
  );
}
