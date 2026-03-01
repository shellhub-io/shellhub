import { useState, FormEvent, useEffect } from "react";
import {
  ShieldCheckIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { getVaultBackend } from "@/utils/vault-backend-factory";
import { INPUT } from "@/utils/styles";

interface Props {
  open: boolean;
  onClose: () => void;
}

export default function VaultSetupDialog({ open, onClose }: Props) {
  const { loading, error, initialize } = useVaultStore();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [legacyCount, setLegacyCount] = useState(0);

  useEffect(() => {
    if (open) {
      setPassword("");
      setConfirm("");
      setLegacyCount(getVaultBackend().loadLegacyKeys().length);
    }
  }, [open]);

  if (!open) return null;

  const passwordTooShort = password.length > 0 && password.length < 8;
  const passwordsMismatch = confirm.length > 0 && password !== confirm;
  const canSubmit =
    password.length >= 8 && password === confirm && !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    await initialize(password);
    if (!useVaultStore.getState().error) {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-[70] flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby="vault-setup-title"
    >
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
            <ShieldCheckIcon className="w-5 h-5 text-primary" />
          </div>
          <div>
            <h2
              id="vault-setup-title"
              className="text-base font-semibold text-text-primary"
            >
              Set Up Secure Vault
            </h2>
            <p className="text-2xs text-text-muted mt-0.5">
              Encrypt your private keys with a master password
            </p>
          </div>
        </div>

        <p className="text-sm text-text-secondary mb-5">
          Your master password protects all stored SSH keys. It cannot be
          recovered — if you forget it, you must reset the vault and lose all
          keys.
        </p>

        {legacyCount > 0 && (
          <div className="flex items-start gap-2.5 bg-accent-yellow/[0.08] border border-accent-yellow/20 rounded-lg px-3.5 py-3 mb-5">
            <ExclamationTriangleIcon className="w-4 h-4 text-accent-yellow shrink-0 mt-0.5" />
            <p className="text-xs text-text-secondary">
              <strong className="text-text-primary">{legacyCount}</strong>{" "}
              existing {legacyCount === 1 ? "key" : "keys"} will be imported
              and encrypted.
            </p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="vault-password"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
            >
              Master Password
            </label>
            <input
              id="vault-password"
              type="password"
              autoComplete="off"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Minimum 8 characters"
              autoFocus
              aria-invalid={passwordTooShort}
              aria-describedby={passwordTooShort ? "vault-password-error" : undefined}
              className={INPUT}
            />
            {passwordTooShort && (
              <p id="vault-password-error" className="text-2xs text-accent-red mt-1.5">
                Password must be at least 8 characters
              </p>
            )}
          </div>

          <div>
            <label
              htmlFor="vault-confirm"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
            >
              Confirm Password
            </label>
            <input
              id="vault-confirm"
              type="password"
              autoComplete="off"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              placeholder="Re-enter your password"
              aria-invalid={passwordsMismatch}
              aria-describedby={passwordsMismatch ? "vault-confirm-error" : undefined}
              className={INPUT}
            />
            {passwordsMismatch && (
              <p id="vault-confirm-error" className="text-2xs text-accent-red mt-1.5">
                Passwords do not match
              </p>
            )}
          </div>

          {error && (
            <p className="text-xs text-accent-red">{error}</p>
          )}

          <div className="flex justify-end gap-2 pt-2">
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
              Create Vault
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
