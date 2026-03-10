import { useState, useEffect, useId, FormEvent } from "react";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { INPUT } from "@/utils/styles";
import BaseDialog from "@/components/common/BaseDialog";

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

  const errorId = `vault-unlock-error-${instanceId}`;

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

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor={`${instanceId}-password`}
            className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
          >
            Master Password
          </label>
          <input
            id={`${instanceId}-password`}
            type="password"
            autoComplete="off"
            data-1p-ignore
            data-lpignore="true"
            data-form-type="other"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter your master password"
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
            className={INPUT}
          />
          {error && (
            <p id={errorId} role="alert" className="text-2xs text-accent-red mt-1.5">
              {error}
            </p>
          )}
        </div>

        <div className="flex items-center justify-between pt-2">
          {onReset ? (
            <button
              type="button"
              onClick={onReset}
              className="text-2xs text-text-muted hover:text-accent-red transition-colors"
            >
              Forgot password? Reset vault
            </button>
          ) : (
            <div />
          )}
          <div className="flex gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canSubmit}
              className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {loading && (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" aria-hidden="true" />
              )}
              Unlock
            </button>
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
