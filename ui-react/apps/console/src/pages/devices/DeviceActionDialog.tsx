import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useAcceptDevice, useRejectDevice, useRemoveDevice } from "@/hooks/useDeviceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";

interface ActionDevice {
  uid: string;
  name: string;
}

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
  open,
}: {
  device: ActionDevice | null;
  action: "accept" | "reject" | "remove";
  onClose: () => void;
  onSuccess?: () => void;
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
        await rejectMutation.mutateAsync({ path: { uid: device.uid, status: "reject" } });
      } else {
        await removeMutation.mutateAsync({ path: { uid: device.uid } });
      }
    } catch (err: unknown) {
      const status = getErrorStatus(err);
      if (action === "accept" && status === 402) {
        setError(
          "Couldn't accept the device. Check your billing status and try again.",
        );
      } else if (action === "accept" && status === 403) {
        setError(
          "You reached the maximum amount of accepted devices in this namespace.",
        );
      } else if (action === "accept" && status === 409) {
        setError(
          "A device with that name already exists. Rename it and try again.",
        );
      } else {
        setError(`Failed to ${action} device.`);
      }
      return;
    }
    onSuccess?.();
    onClose();
  };

  const description = device
    ? (
      <>
        {config.description}
        {" "}
        <span className="font-medium text-text-primary">{device.name}</span>
        ?
        {action === "remove" && (
          <p className="text-xs text-text-muted/70 mt-1">
            This action cannot be undone.
          </p>
        )}
      </>
    )
    : null;

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleConfirm}
      title={config.title}
      description={description}
      confirmLabel={config.confirm}
      variant={config.variant}
    >
      {error && (
        <p role="alert" className="text-xs font-mono text-accent-red mb-2 flex items-center gap-1.5">
          <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
          {error}
        </p>
      )}
    </ConfirmDialog>
  );
}

export default DeviceActionDialog;
