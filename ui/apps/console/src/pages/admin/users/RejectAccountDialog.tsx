import { useState } from "react";
import { NoSymbolIcon } from "@heroicons/react/24/outline";
import { useDeleteUser } from "@/hooks/useAdminUserMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";

interface RejectAccountDialogProps {
  open: boolean;
  onClose: () => void;
  user: { id: string; email: string } | null;
}

/**
 * Confirms rejecting an account request, which deletes the pending account.
 */
export default function RejectAccountDialog({
  open,
  onClose,
  user,
}: RejectAccountDialogProps) {
  const deleteUser = useDeleteUser();
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open={open}
      onClose={() => {
        setError("");
        onClose();
      }}
      onConfirm={async () => {
        if (!user) return;
        setError("");
        try {
          await deleteUser.mutateAsync({ path: { id: user.id } });
          onClose();
        } catch {
          setError("Failed to reject the account. Please try again.");
        }
      }}
      icon={<NoSymbolIcon />}
      title="Reject user"
      description={
        <>
          <ObjectName>{user?.email}</ObjectName> is removed and doesn&apos;t
          join the namespace that added them.
        </>
      }
      errorMessage={error || null}
      confirmLabel="Reject"
    />
  );
}
