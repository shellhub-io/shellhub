import { useState } from "react";
import { useDeleteUser } from "../../../hooks/useAdminUserMutations";
import ConfirmDialog from "../../../components/common/ConfirmDialog";

interface DeleteUserDialogProps {
  open: boolean;
  onClose: () => void;
  user: { id: string; name: string } | null;
  onDeleted?: () => void;
}

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
      title="Delete User"
      description={(
        <>
          Are you sure you want to remove{" "}
          <span className="font-medium text-text-primary">{user?.name}</span>{" "}
          and all associated namespace data? This action cannot be undone.
          {error && (
            <span className="block mt-2 text-accent-red text-2xs">{error}</span>
          )}
        </>
      )}
      confirmLabel="Delete"
    />
  );
}
