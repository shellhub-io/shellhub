import { useState, useEffect, useId, FormEvent } from "react";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import BaseDialog from "@/components/common/BaseDialog";
import PasswordField from "@/components/common/fields/PasswordField";
import { Button } from "@shellhub/design-system/primitives";

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
    if (!useVaultStore.getState().error && !useVaultStore.getState().loading) {
      onClose();
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
          <LockClosedIcon className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h2
            id={`vault-unlock-title-${instanceId}`}
            className="text-base font-semibold text-text-primary"
          >
            Unlock Vault
          </h2>
          <p className="text-2xs text-text-muted mt-0.5">
            Enter your master password to access your keys
          </p>
        </div>
      </div>

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
  );
}

export default function VaultUnlockDialog({ open, onClose, onReset }: Props) {
  const instanceId = useId();
  const titleId = `vault-unlock-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
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
