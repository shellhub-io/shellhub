import { useState } from "react";
import { useAdminDeleteNamespace } from "@/hooks/useAdminNamespaceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";

interface DeleteNamespaceDialogProps {
  open: boolean;
  onClose: () => void;
  namespace: { tenant_id: string; name: string } | null;
  onDeleted?: () => void;
}

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
      title="Delete Namespace"
      description={(
        <>
          Are you sure you want to delete{" "}
          <span className="font-medium text-text-primary">
            {namespace?.name}
          </span>
          ? This will permanently remove all devices, sessions, public keys, and
          API keys associated with this namespace. This action cannot be undone.
          {error && (
            <span className="block mt-2 text-accent-red text-2xs">{error}</span>
          )}
        </>
      )}
      confirmLabel="Delete"
    />
  );
}
