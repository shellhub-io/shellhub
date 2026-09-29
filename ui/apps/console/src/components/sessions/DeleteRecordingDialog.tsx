import { useState } from "react";
import { TrashIcon } from "@heroicons/react/24/outline";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { useRemoveRecording } from "@/hooks/useRemoveRecording";

/**
 * Confirms, then deletes every copy of a session's recording, the server's included when recorded
 * is set. It stays open with the error shown when the delete fails, and closes when it succeeds.
 */
export default function DeleteRecordingDialog({
  open,
  onClose,
  sessionUid,
  recorded,
}: {
  open: boolean;
  onClose: () => void;
  sessionUid: string;
  recorded: boolean;
}) {
  const remove = useRemoveRecording();
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    setError(null);
    try {
      await remove(sessionUid, recorded);
      onClose();
    } catch {
      setError("Failed to delete recording.");
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
