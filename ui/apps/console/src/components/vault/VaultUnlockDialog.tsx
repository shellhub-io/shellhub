import { useState, useEffect, useId, FormEvent } from "react";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import BaseDialog from "@/components/common/BaseDialog";
import PasswordField from "@/components/common/fields/PasswordField";
import { Button } from "@shellhub/design-system/primitives";
import DialogHeader from "@/components/common/DialogHeader";

interface Props {
  open: boolean;
  onClose: () => void;
  onReset?: () => void;
}

interface FormProps extends Props {
  instanceId: string;
}

function UnlockForm({ open, onClose, onReset, instanceId }: FormProps) {
  const loading = useVaultStore((s) => s.loading);
  const error = useVaultStore((s) => s.error);
  const unlock = useVaultStore((s) => s.unlock);
  const clearError = useVaultStore((s) => s.clearError);
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (open) clearError();
  }, [open, clearError]);

  const canSubmit = password.length > 0 && !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    await unlock(password);
    if (useVaultStore.getState().status === "unlocked") {
      onClose();
    }
  };

  return (
    <div>
      <DialogHeader
        icon={<LockClosedIcon />}
        title="Unlock vault"
        description="Enter your master password to reach your keys."
        titleId={`vault-unlock-title-${instanceId}`}
        descriptionId={`vault-unlock-title-${instanceId}-description`}
        onClose={onClose}
      />
      <div className="px-6 pb-6">

      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
        <PasswordField
          id={`${instanceId}-password`}
          label="Master Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter your master password"
          suppressPasswordManager
          error={error ?? undefined}
          errorRole="alert"
        />

        <div className="flex items-center justify-between pt-2">
          {onReset ? (
            <Button variant="ghost" size="sm" onClick={onReset}>
              Forgot password? Reset vault
            </Button>
          ) : (
            <div />
          )}
          <div className="flex gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={!canSubmit} loading={loading}>
              Unlock
            </Button>
          </div>
        </div>
      </form>
      </div>
    </div>
  );
}

/**
 * Asks for the vault passphrase. onReset is the way out for someone who has lost it, and it
 * destroys the keys rather than recovering them.
 */
export default function VaultUnlockDialog({ open, onClose, onReset }: Props) {
  const instanceId = useId();
  const titleId = `vault-unlock-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={`${titleId}-description`}
    >
      <UnlockForm
        key={String(open)}
        open={open}
        onClose={onClose}
        onReset={onReset}
        instanceId={instanceId}
      />
    </BaseDialog>
  );
}
