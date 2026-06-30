import { useState } from "react";
import { isSdkError } from "@/api/errors";
import {
  useAcceptDevice,
  useRejectDevice,
  useRemoveDevice,
} from "@/hooks/useDeviceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { getAcceptDeviceErrorMessage } from "@/utils/deviceErrors";
import type { EntityBase } from "@/hooks/useActionDialogState";

const ACTION_CONFIG = {
  accept: {
    title: "Accept Device",
    description: "Do you want to accept",
    confirm: "Accept",
    variant: "success" as const,
  },
  reject: {
    title: "Reject Device",
    description: "Do you want to reject",
    confirm: "Reject",
    variant: "warning" as const,
  },
  remove: {
    title: "Remove Device",
    description: "Do you want to remove",
    confirm: "Remove",
    variant: "danger" as const,
  },
};

function getErrorStatus(err: unknown): number | undefined {
  return isSdkError(err) ? err.status : undefined;
}

function DeviceActionDialog({
  device,
  action,
  onClose,
  onSuccess,
  onBillingWarning,
  open,
}: {
  device: EntityBase | null;
  action: "accept" | "reject" | "remove";
  onClose: () => void;
  onSuccess?: () => void;
  onBillingWarning?: () => void;
  open: boolean;
}) {
  const acceptMutation = useAcceptDevice();
  const rejectMutation = useRejectDevice();
  const removeMutation = useRemoveDevice();
  const [error, setError] = useState<string | null>(null);
  const config = ACTION_CONFIG[action];

  const handleConfirm = async () => {
    if (!device) return;
    setError(null);
    try {
      if (action === "accept") {
        await acceptMutation.mutateAsync({ path: { uid: device.uid } });
      } else if (action === "reject") {
        await rejectMutation.mutateAsync({
          path: { uid: device.uid, status: "reject" },
        });
      } else {
        await removeMutation.mutateAsync({ path: { uid: device.uid } });
      }
    } catch (err: unknown) {
      const status = getErrorStatus(err);
      if (action !== "accept") {
        setError(`Failed to ${action} device.`);
        return;
      }

      if (status === 402 && onBillingWarning) {
        onBillingWarning();
        return;
      }

      setError(getAcceptDeviceErrorMessage(err));
      return;
    }
    onSuccess?.();
    onClose();
  };

  const description = device ? (
    <>
      {config.description}{" "}
      <span className="font-medium text-text-primary">{device.name}</span>?
      {action === "remove" && (
        <p className="text-xs text-text-muted/70 mt-1">
          This action cannot be undone.
        </p>
      )}
    </>
  ) : null;

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleConfirm}
      title={config.title}
      description={description}
      confirmLabel={config.confirm}
      variant={config.variant}
      errorMessage={error}
    />
  );
}

export default DeviceActionDialog;
