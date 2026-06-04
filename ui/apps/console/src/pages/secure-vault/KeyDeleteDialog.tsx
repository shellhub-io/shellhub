import { useState } from "react";
import { useVaultStore } from "@/stores/vaultStore";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import type { VaultKeyEntry } from "@/types/vault";

interface Props {
  open: boolean;
  entry: VaultKeyEntry | null;
  onClose: () => void;
}

export default function KeyDeleteDialog({ open, entry, onClose }: Props) {
  const removeKey = useVaultStore((s) => s.removeKey);
  const [error, setError] = useState<string | null>(null);

  const handleConfirm = async () => {
    if (!entry) return;
    setError(null);
    try {
      await removeKey(entry.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete key.");
    }
  };

  const handleClose = () => {
    setError(null);
    onClose();
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={handleClose}
      onConfirm={handleConfirm}
      title="Delete Private Key"
      description={(
        <>
          Are you sure you want to delete
          {" "}
          <strong className="text-text-primary">{entry?.name}</strong>
          ? This
          action cannot be undone.
        </>
      )}
      confirmLabel="Delete"
      variant="danger"
    >
      {error && (
        <p className="text-xs text-accent-red">{error}</p>
      )}
    </ConfirmDialog>
  );
}
