import { useState } from "react";
import { isSdkError } from "@/api/errors";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import {
  useUpdateContainerStatus,
  useRemoveContainer,
} from "@/hooks/useContainerMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";

interface ActionContainer {
  uid: string;
  name: string;
}

const ACTION_CONFIG = {
  accept: {
    title: "Accept Container",
    description: "Do you want to accept",
    confirm: "Accept",
    variant: "success" as const,
  },
  reject: {
    title: "Reject Container",
    description: "Do you want to reject",
    confirm: "Reject",
    variant: "warning" as const,
  },
  remove: {
    title: "Remove Container",
    description: "Do you want to remove",
    confirm: "Remove",
    variant: "danger" as const,
  },
};

function getErrorStatus(err: unknown): number | undefined {
  return isSdkError(err) ? err.status : undefined;
}

function ContainerActionDialog({
  container,
  action,
  onClose,
  onSuccess,
  onBillingWarning,
  open,
}: {
  container: ActionContainer | null;
  action: "accept" | "reject" | "remove";
  onClose: () => void;
  onSuccess?: () => void;
  onBillingWarning?: () => void;
  open: boolean;
}) {
  const statusMutation = useUpdateContainerStatus();
  const removeMutation = useRemoveContainer();
  const [error, setError] = useState<string | null>(null);
  const config = ACTION_CONFIG[action];

  const handleConfirm = async () => {
    if (!container) return;
    setError(null);
    try {
      if (action === "accept" || action === "reject") {
        await statusMutation.mutateAsync({
          path: { uid: container.uid, status: action },
        });
      } else {
        await removeMutation.mutateAsync({ path: { uid: container.uid } });
      }
    } catch (err: unknown) {
      const status = getErrorStatus(err);
      if (action === "accept" && status === 402 && onBillingWarning) {
        onBillingWarning();
        return;
      }
      if (action === "accept" && status === 402) {
        setError(
          "Couldn't accept the container. Check your billing status and try again.",
        );
      } else if (action === "accept" && status === 403) {
        setError(
          "You reached the maximum amount of accepted containers in this namespace.",
        );
      } else if (action === "accept" && status === 409) {
        setError(
          "A container with that name already exists. Rename it and try again.",
        );
      } else {
        setError(`Failed to ${action} container.`);
      }
      return;
    }
    onSuccess?.();
    onClose();
  };

  const description = container ? (
    <>
      {config.description}{" "}
      <span className="font-medium text-text-primary">{container.name}</span>?
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
    >
      {error && (
        <p
          role="alert"
          className="text-xs font-mono text-accent-red mb-2 flex items-center gap-1.5"
        >
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error}
        </p>
      )}
    </ConfirmDialog>
  );
}

export default ContainerActionDialog;
