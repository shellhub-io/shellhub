import { useState } from "react";
import {
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAdminDeleteAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";

interface DeleteAnnouncementDialogProps {
  open: boolean;
  onClose: () => void;
  announcement: { uuid: string; title: string } | null;
  onDeleted?: () => void;
}

/**
 * Confirms deleting an announcement. It is already visible to users, so this names it rather
 * than asking a generic question.
 */
export default function DeleteAnnouncementDialog({
  open,
  onClose,
  announcement,
  onDeleted,
}: DeleteAnnouncementDialogProps) {
  const deleteAnnouncement = useAdminDeleteAnnouncement();
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open={open}
      onClose={() => {
        setError("");
        onClose();
      }}
      onConfirm={async () => {
        if (!announcement) return;
        setError("");
        try {
          await deleteAnnouncement.mutateAsync({
            path: { uuid: announcement.uuid },
          });
          onClose();
          onDeleted?.();
        } catch {
          setError("Failed to delete announcement. Please try again.");
        }
      }}
      icon={<TrashIcon />}
      title="Delete announcement"
      description={
        <>
          <ObjectName>{announcement?.title}</ObjectName> stops showing to users.
          This can't be undone.
        </>
      }
      errorMessage={error || null}
      confirmLabel="Delete announcement"
    />
  );
}
