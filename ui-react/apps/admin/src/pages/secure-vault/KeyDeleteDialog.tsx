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

  const handleConfirm = async () => {
    if (!entry) return;
    await removeKey(entry.id);
    onClose();
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleConfirm}
      title="Delete Private Key"
      description={
        <>
          Are you sure you want to delete{" "}
          <strong className="text-text-primary">{entry?.name}</strong>? This
          action cannot be undone.
        </>
      }
      confirmLabel="Delete"
      variant="danger"
    />
  );
}
