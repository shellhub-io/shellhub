import { useCallback } from "react";
import {
  useAcceptDevice,
  useRejectDevice,
  useRemoveDevice,
} from "@/hooks/useDeviceMutations";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

export function useDeviceActionRunner() {
  const accept = useAcceptDevice();
  const reject = useRejectDevice();
  const remove = useRemoveDevice();

  return useCallback(
    async (entity: EntityBase, operation: EntityOperation) => {
      if (operation === "reject") {
        await reject.mutateAsync({ path: { uid: entity.uid, status: "reject" } });
        return;
      }

      const action = operation === "accept" ? accept : remove;
      await action.mutateAsync({ path: { uid: entity.uid } });
    },
    [accept, reject, remove],
  );
}
