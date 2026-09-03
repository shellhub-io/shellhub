import { useState } from "react";
import { useInstallKeyUpdate } from "@/client/api";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { type InstallKey } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import InputField from "@/components/common/fields/InputField";

/**
 * Confirms revoking an install key. Revoking is final — unlike disabling, it cannot be undone —
 * which is what the dialog has to make clear.
 */
export default function RevokeInstallKeyDialog({
  installKey,
  onRevoked,
}: {
  installKey: InstallKey | null;
  onRevoked: () => void;
}) {
  const updateKey = useInstallKeyUpdate();
  const [confirmText, setConfirmText] = useState("");
  const [error, setError] = useState<string | null>(null);

  const open = !!installKey;
  const name = installKey?.name ?? "";

  useResetOnOpen(open, () => {
    setConfirmText("");
    setError(null);
  });

  const handleConfirm = async () => {
    if (!installKey) return;
    setError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: installKey.name },
        body: { revoked: true },
      });
      onRevoked();
    } catch {
      setError("Failed to revoke Install Key.");
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onRevoked}
      onConfirm={handleConfirm}
      title="Revoke Install Key"
      description={
        <>
          Revoking <span className="font-medium text-text-primary">{name}</span>{" "}
          is permanent — there's no undo. Any device or pipeline still using
          this key to register will stop; devices already registered keep
          working. Type{" "}
          <code className="font-mono text-accent-red">{name}</code> to confirm.
        </>
      }
      confirmLabel="Revoke key"
      confirmDisabled={confirmText !== name}
      errorMessage={error}
    >
      <InputField
        id="revoke-install-key-confirm"
        label="Type the key's name to confirm"
        hideLabel
        value={confirmText}
        onChange={setConfirmText}
        autoComplete="off"
      />
    </ConfirmDialog>
  );
}
