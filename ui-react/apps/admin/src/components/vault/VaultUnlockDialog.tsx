import { useState, useEffect, FormEvent } from "react";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { INPUT } from "@/utils/styles";

interface Props {
  open: boolean;
  onClose: () => void;
  onReset?: () => void;
}

export default function VaultUnlockDialog({ open, onClose, onReset }: Props) {
  const { loading, error, unlock } = useVaultStore();
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (open) setPassword("");
  }, [open]);

  if (!open) return null;

  const canSubmit = password.length > 0 && !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    await unlock(password);
    if (!useVaultStore.getState().error) {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-[70] flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby="vault-unlock-title"
    >
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
            <LockClosedIcon className="w-5 h-5 text-primary" />
          </div>
          <div>
            <h2
              id="vault-unlock-title"
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
              htmlFor="vault-unlock-password"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
            >
              Master Password
            </label>
            <input
              id="vault-unlock-password"
              type="text"
              inputMode="text"
              autoComplete="off"
              data-1p-ignore
              data-lpignore="true"
              data-form-type="other"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your master password"
              autoFocus
              aria-invalid={!!error}
              aria-describedby={error ? "vault-unlock-error" : undefined}
              className={`${INPUT} [-webkit-text-security:disc]`}
            />
            {error && (
              <p id="vault-unlock-error" className="text-2xs text-accent-red mt-1.5">
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
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                )}
                Unlock
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
