import { useMutation } from "@tanstack/react-query";
import {
  deleteContainerMutation,
  updateContainerStatusMutation,
} from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/**
 * Runs the confirmation dialog's chosen operation against a container, so the dialog does not
 * have to know which mutation each operation maps to.
 */
export function useContainerActionRunner() {
  const invalidateContainers = useInvalidateByIds("getContainers", "getContainer");
  const status = useMutation({
    ...updateContainerStatusMutation(),
    onSuccess: invalidateContainers,
  });
  const remove = useMutation({
    ...deleteContainerMutation(),
    onSuccess: invalidateContainers,
  });

  return async (entity: EntityBase, operation: EntityOperation) => {
    if (operation === "remove") {
      await remove.mutateAsync({ path: { uid: entity.uid } });
      return;
    }

    await status.mutateAsync({ path: { uid: entity.uid, status: operation } });
  };
}
