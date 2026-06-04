import { useState, FormEvent, useEffect, useId, useMemo } from "react";
import {
  ShieldCheckIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { getVaultBackend } from "@/utils/vault-backend-factory";
import { useAuthStore } from "@/stores/authStore";
import BaseDialog from "@/components/common/BaseDialog";
import PasswordField from "@/components/common/fields/PasswordField";
import Spinner from "@/components/common/Spinner";

interface Props {
  open: boolean;
  onClose: () => void;
}

interface FormProps extends Props {
  instanceId: string;
}

function SetupForm({ open, onClose, instanceId }: FormProps) {
  const loading = useVaultStore((s) => s.loading);
  const error = useVaultStore((s) => s.error);
  const initialize = useVaultStore((s) => s.initialize);
  const clearError = useVaultStore((s) => s.clearError);
  const user = useAuthStore((s) => s.user);
  const tenant = useAuthStore((s) => s.tenant);
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");

  const titleId = `vault-setup-title-${instanceId}`;

  useEffect(() => {
    if (open) clearError();
  }, [open, clearError]);

  // Compute once per open/user/tenant change — avoids a localStorage read
  // on every render and ensures the component subscribes to auth changes.
  const legacyCount = useMemo(() => {
    if (!open) return 0;
    const scope = user && tenant ? { user, tenant } : undefined;
    return getVaultBackend(scope).loadLegacyKeys().length;
  }, [open, user, tenant]);

  const passwordTooShort = password.length > 0 && password.length < 8;
  const passwordsMismatch = confirm.length > 0 && password !== confirm;
  const canSubmit = password.length >= 8 && password === confirm && !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    await initialize(password);
    if (!useVaultStore.getState().error && !useVaultStore.getState().loading) {
      onClose();
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
          <ShieldCheckIcon className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h2
            id={titleId}
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
            existing {legacyCount === 1 ? "key" : "keys"} will be imported and
            encrypted.
          </p>
        </div>
      )}

      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
        <PasswordField
          id={`${instanceId}-password`}
          label="Master Password"
          value={password}
          onChange={setPassword}
          placeholder="Minimum 8 characters"
          suppressPasswordManager
          error={
            passwordTooShort
              ? "Password must be at least 8 characters"
              : undefined
          }
        />

        <PasswordField
          id={`${instanceId}-confirm`}
          label="Confirm Password"
          value={confirm}
          onChange={setConfirm}
          placeholder="Re-enter your password"
          suppressPasswordManager
          error={passwordsMismatch ? "Passwords do not match" : undefined}
        />

        {error && (
          <p role="alert" className="text-xs text-accent-red">
            {error}
          </p>
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
              <Spinner tone="onPrimary" />
            )}
            Create Vault
          </button>
        </div>
      </form>
    </div>
  );
}

export default function VaultSetupDialog({ open, onClose }: Props) {
  const instanceId = useId();
  const titleId = `vault-setup-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
    >
      <SetupForm
        key={String(open)}
        open={open}
        onClose={onClose}
        instanceId={instanceId}
      />
    </BaseDialog>
  );
}
