import { useState } from "react";
import {
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useDeleteUser } from "@/hooks/useAdminUserMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";

interface DeleteUserDialogProps {
  open: boolean;
  onClose: () => void;
  user: { id: string; name: string } | null;
  onDeleted?: () => void;
}

/**
 * Confirms deleting a user, naming them so the wrong row cannot be confirmed by habit.
 */
export default function DeleteUserDialog({
  open,
  onClose,
  user,
  onDeleted,
}: DeleteUserDialogProps) {
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
          onDeleted?.();
        } catch {
          setError("Failed to delete user. Please try again.");
        }
      }}
      icon={<TrashIcon />}
      title="Delete user"
      description={
        <>
          <ObjectName>{user?.name}</ObjectName> is removed with all the
          namespace data they own. This can't be undone.
        </>
      }
      errorMessage={error || null}
      confirmLabel="Delete user"
    />
  );
}
