import { useCallback } from "react";
import { useUpdateContainerStatus, useDeleteContainer } from "@/client/api";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/** Maps action-dialog operations to container mutations. */
export function useContainerActionRunner() {
  const status = useUpdateContainerStatus();
  const remove = useDeleteContainer();

  return useCallback(
    async (entity: EntityBase, operation: EntityOperation) => {
      if (operation === "remove") {
        await remove.mutateAsync({ uid: entity.uid });
        return;
      }

      await status.mutateAsync({ uid: entity.uid, status: operation });
    },
    [status, remove],
  );
}
