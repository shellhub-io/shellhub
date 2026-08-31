import { useCallback } from "react";
import {
  useUpdateContainerStatus,
  useRemoveContainer,
} from "@/hooks/useContainerMutations";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/**
 * Runs the confirmation dialog's chosen operation against a container, so the dialog does not
 * have to know which mutation each operation maps to.
 */
export function useContainerActionRunner() {
  const status = useUpdateContainerStatus();
  const remove = useRemoveContainer();

  return useCallback(
    async (entity: EntityBase, operation: EntityOperation) => {
      if (operation === "remove") {
        await remove.mutateAsync({ path: { uid: entity.uid } });
        return;
      }

      await status.mutateAsync({ path: { uid: entity.uid, status: operation } });
    },
    [status, remove],
  );
}
