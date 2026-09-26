import { useState } from "react";
import {
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAdminDeleteNamespace } from "@/hooks/useAdminNamespaceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";

interface DeleteNamespaceDialogProps {
  open: boolean;
  onClose: () => void;
  namespace: { tenant_id: string; name: string } | null;
  onDeleted?: () => void;
}

/**
 * Confirms deleting a namespace. Everything in it goes too, so the dialog says so and names the
 * namespace rather than asking a generic question.
 */
export default function DeleteNamespaceDialog({
  open,
  onClose,
  namespace,
  onDeleted,
}: DeleteNamespaceDialogProps) {
  const deleteNamespace = useAdminDeleteNamespace();
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open={open}
      onClose={() => {
        setError("");
        onClose();
      }}
      onConfirm={async () => {
        if (!namespace) return;
        setError("");
        try {
          await deleteNamespace.mutateAsync({
            path: { tenant: namespace.tenant_id },
          });
          onClose();
          onDeleted?.();
        } catch {
          setError("Failed to delete namespace. Please try again.");
        }
      }}
      icon={<TrashIcon />}
      title="Delete namespace"
      description={
        <>
          <ObjectName>{namespace?.name}</ObjectName> goes away with its devices,
          sessions, public keys and API keys. This can't be undone.
        </>
      }
      errorMessage={error || null}
      confirmLabel="Delete namespace"
    />
  );
}
