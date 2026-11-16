import { useMutation } from "@tanstack/react-query";
import { deleteDeviceMutation, updateDeviceStatusMutation } from "@/client";
import { useAcceptDevice } from "@/hooks/useDeviceRequests";
import {
  useInvalidateByIds,
  DEVICE_STATUS_QUERIES,
} from "@/hooks/useInvalidateQueries";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";

/**
 * Runs the confirmation dialog's chosen operation against a device, so the dialog does not have
 * to know which mutation each operation maps to.
 */
export function useDeviceActionRunner() {
  const accept = useAcceptDevice();
  const invalidateDevices = useInvalidateByIds(...DEVICE_STATUS_QUERIES);
  const reject = useMutation({
    ...updateDeviceStatusMutation(),
    onSuccess: invalidateDevices,
  });
  const remove = useMutation({
    ...deleteDeviceMutation(),
    onSuccess: invalidateDevices,
  });

  return async (
    entity: EntityBase,
    operation: EntityOperation,
    replace = false,
  ) => {
    if (operation === "reject") {
      await reject.mutateAsync({ path: { uid: entity.uid, status: "reject" } });
      return;
    }

    if (operation === "accept") {
      await accept.mutateAsync({
        path: { uid: entity.uid },
        ...(replace ? { body: { replace: true } } : {}),
      });
      return;
    }

    await remove.mutateAsync({ path: { uid: entity.uid } });
  };
}
