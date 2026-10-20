import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import {
  TrashIcon,
} from "@heroicons/react/24/outline";
import { isSdkError } from "@/api/errors";
import { isCloud } from "@/env";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";
import { deleteNamespaceAdminMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

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
  const invalidateNamespaces = useInvalidateByIds(
    "getNamespacesAdmin",
    "getNamespaceAdmin",
  );
  const deleteNamespace = useMutation({
    ...deleteNamespaceAdminMutation(),
    onSuccess: invalidateNamespaces,
  });
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
        } catch (err) {
          setError(
            isCloud() && isSdkError(err) && err.status === 402
              ? "This namespace has an active subscription or an unpaid invoice, which blocks its deletion."
              : "Failed to delete namespace. Please try again.",
          );
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
