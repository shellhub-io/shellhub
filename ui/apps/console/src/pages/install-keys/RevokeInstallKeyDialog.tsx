import ConfirmDialog from "@/components/common/ConfirmDialog";
import InputField from "@/components/common/fields/InputField";
import { type InstallKey } from "@/client";

/**
 * The type-to-confirm revoke dialog for an install key, shared by the list page (one dialog for the
 * targeted row) and the activity-page actions menu (bound to its single key). The caller owns the
 * open/confirm-text/error state since the two drive it differently.
 */
export default function RevokeInstallKeyDialog({
  installKey,
  open,
  confirmText,
  onConfirmTextChange,
  onClose,
  onConfirm,
  error,
}: {
  installKey: InstallKey | null;
  open: boolean;
  confirmText: string;
  onConfirmTextChange: (value: string) => void;
  onClose: () => void;
  onConfirm: () => Promise<void> | void;
  error: string | null;
}) {
  const name = installKey?.name ?? "";

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={onConfirm}
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
        onChange={onConfirmTextChange}
        autoComplete="off"
      />
    </ConfirmDialog>
  );
}
