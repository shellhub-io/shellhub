import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { TrashIcon } from "@heroicons/react/24/outline";
import { webhookDeleteMutation, type Webhook } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";

/**
 * Confirms deleting a webhook. Its pending deliveries go with it and are never sent.
 */
export default function DeleteWebhookDialog({
  webhook,
  onClose,
}: {
  webhook: Webhook | null;
  onClose: () => void;
}) {
  const invalidate = useInvalidateByIds("webhookList");
  const deleteWebhook = useMutation({ ...webhookDeleteMutation(), onSuccess: invalidate });
  const [error, setError] = useState<string | null>(null);
  const open = !!webhook;

  useResetOnOpen(open, () => setError(null));

  const handleConfirm = async () => {
    if (!webhook) return;
    setError(null);
    try {
      await deleteWebhook.mutateAsync({ path: { id: webhook.id } });
      onClose();
    } catch {
      setError("Couldn't delete the webhook.");
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleConfirm}
      variant="danger"
      icon={<TrashIcon />}
      title="Delete webhook"
      description={
        <>
          ShellHub stops sending events to <ObjectName>{webhook?.url ?? ""}</ObjectName>, and
          deliveries still waiting to be sent are dropped.
        </>
      }
      confirmLabel="Delete webhook"
      errorMessage={error}
    />
  );
}
