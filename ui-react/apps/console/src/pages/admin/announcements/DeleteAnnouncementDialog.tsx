import { useState } from "react";
import { useAdminDeleteAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";

interface DeleteAnnouncementDialogProps {
  open: boolean;
  onClose: () => void;
  announcement: { uuid: string; title: string } | null;
  onDeleted?: () => void;
}

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
      title="Delete Announcement"
      description={
        <>
          Are you sure you want to delete{" "}
          <span className="font-medium text-text-primary">
            {announcement?.title}
          </span>
          ? This action cannot be undone.
          {error && (
            <span className="block mt-2 text-accent-red text-2xs">{error}</span>
          )}
        </>
      }
      confirmLabel="Delete"
    />
  );
}
