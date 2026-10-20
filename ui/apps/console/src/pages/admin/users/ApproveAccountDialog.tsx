import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { CheckBadgeIcon } from "@heroicons/react/24/outline";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";
import { approveUserMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

interface ApproveAccountDialogProps {
  open: boolean;
  onClose: () => void;
  user: { id: string; email: string } | null;
}

/**
 * Confirms approving an account request. The account is already confirmed when it awaits
 * approval, so approving it is all that stands between the person and signing in.
 */
export default function ApproveAccountDialog({
  open,
  onClose,
  user,
}: ApproveAccountDialogProps) {
  const invalidateApproved = useInvalidateByIds(
    "getUsers",
    "getUser",
    "getNamespaceAdmin",
  );
  const approve = useMutation({
    ...approveUserMutation(),
    onSuccess: invalidateApproved,
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
        if (!user) return;
        setError("");
        try {
          await approve.mutateAsync({ path: { id: user.id } });
          onClose();
        } catch {
          setError("Failed to approve the account. Please try again.");
        }
      }}
      icon={<CheckBadgeIcon />}
      title="Approve user"
      description={
        <>
          <ObjectName>{user?.email}</ObjectName> can sign in as soon as you
          approve.
        </>
      }
      errorMessage={error || null}
      confirmLabel="Approve"
      variant="primary"
    />
  );
}
