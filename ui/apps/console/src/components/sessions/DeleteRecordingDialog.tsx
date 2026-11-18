import { useState } from "react";
import { TrashIcon } from "@heroicons/react/24/outline";
import { isSdkError } from "@/api/errors";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { useRemoveRecording } from "@/hooks/useRemoveRecording";

/**
 * Confirms, then deletes every copy of the given seats' recordings, seat zero alone by default,
 * the server's included when recorded is set. It stays open with the error shown when the delete
 * fails, with the copies before the failure already gone and onDeleted not called, and closes
 * when it succeeds, calling onDeleted first. The server refuses with 409 while the session is still
 * active, which the error names.
 */
export default function DeleteRecordingDialog({
  open,
  onClose,
  sessionUid,
  recorded,
  seats,
  onDeleted,
}: {
  open: boolean;
  onClose: () => void;
  sessionUid: string;
  recorded: boolean;
  seats?: number[];
  onDeleted?: () => void;
}) {
  const remove = useRemoveRecording();
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    setError(null);
    try {
      await remove(sessionUid, recorded, seats);
      onDeleted?.();
      onClose();
    } catch (err: unknown) {
      setError(
        isSdkError(err) && err.status === 409
          ? "Close the session before deleting its recording."
          : "Failed to delete recording.",
      );
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={() => {
        setError(null);
        onClose();
      }}
      onConfirm={handleDelete}
      icon={<TrashIcon />}
      title="Delete recording"
      description="This session's recording is deleted and can't be played back again."
      confirmLabel="Delete recording"
      variant="danger"
      errorMessage={error}
    />
  );
}
