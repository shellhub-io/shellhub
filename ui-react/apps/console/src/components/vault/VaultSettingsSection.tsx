import { useState, useEffect, FormEvent } from "react";
import {
  KeyIcon,
  LockClosedIcon,
  ExclamationTriangleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import Spinner from "@/components/common/Spinner";

function ChangePasswordDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const loading = useVaultStore((s) => s.loading);
  const error = useVaultStore((s) => s.error);
  const changeMasterPassword = useVaultStore((s) => s.changeMasterPassword);
  const clearError = useVaultStore((s) => s.clearError);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [prevOpen, setPrevOpen] = useState(false);
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (open) {
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    }
  }

  useEffect(() => {
    if (open) clearError();
  }, [open, clearError]);

  const newTooShort = newPassword.length > 0 && newPassword.length < 8;
  const mismatch =
    confirmPassword.length > 0 && newPassword !== confirmPassword;
  const canSubmit =
    currentPassword.length > 0 &&
    newPassword.length >= 8 &&
    newPassword === confirmPassword &&
    !loading;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit) return;
    await changeMasterPassword(currentPassword, newPassword);
    if (!useVaultStore.getState().error) {
      onClose();
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Change Master Password"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {loading && (
              <Spinner size="sm" tone="onPrimary" />
            )}
            Update Password
          </button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <PasswordField
          id="vault-current-password"
          label="Current Password"
          value={currentPassword}
          onChange={setCurrentPassword}
          placeholder="Enter current master password"
          autoFocus={open}
          suppressPasswordManager
        />

        <div className="h-px bg-border" />

        <PasswordField
          id="vault-new-password"
          label="New Password"
          value={newPassword}
          onChange={setNewPassword}
          placeholder="Minimum 8 characters"
          suppressPasswordManager
          error={
            newTooShort ? "Password must be at least 8 characters" : undefined
          }
        />

        <PasswordField
          id="vault-confirm-new-password"
          label="Confirm New Password"
          value={confirmPassword}
          onChange={setConfirmPassword}
          placeholder="Re-enter new password"
          suppressPasswordManager
          error={mismatch ? "Passwords do not match" : undefined}
        />

        {error && (
          <p className="text-xs text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" />
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}

export default function VaultSettingsSection() {
  const status = useVaultStore((s) => s.status);
  const lock = useVaultStore((s) => s.lock);
  const resetVault = useVaultStore((s) => s.resetVault);
  const [changeOpen, setChangeOpen] = useState(false);
  const [resetOpen, setResetOpen] = useState(false);
  const [resetConfirmText, setResetConfirmText] = useState("");

  if (status !== "unlocked") return null;

  return (
    <>
      <div className="mt-8 animate-fade-in">
        <h3 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
          Vault Settings
        </h3>
        <div className="bg-card border border-border rounded-lg divide-y divide-border">
          <button
            onClick={() => setChangeOpen(true)}
            className="flex items-center gap-3 w-full px-4 py-3.5 text-left hover:bg-hover-subtle transition-colors rounded-t-lg"
          >
            <KeyIcon className="w-4 h-4 text-text-muted shrink-0" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-text-primary">
                Change Master Password
              </p>
              <p className="text-2xs text-text-muted">
                Re-encrypt all keys with a new password.
              </p>
            </div>
          </button>

          <button
            onClick={lock}
            className="flex items-center gap-3 w-full px-4 py-3.5 text-left hover:bg-hover-subtle transition-colors"
          >
            <LockClosedIcon className="w-4 h-4 text-text-muted shrink-0" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-text-primary">
                Lock Vault
              </p>
              <p className="text-2xs text-text-muted">
                Clear decrypted keys from memory.
              </p>
            </div>
          </button>

          <button
            onClick={() => {
              setResetConfirmText("");
              setResetOpen(true);
            }}
            className="flex items-center gap-3 w-full px-4 py-3.5 text-left hover:bg-accent-red/5 transition-colors rounded-b-lg"
          >
            <ExclamationTriangleIcon className="w-4 h-4 text-accent-red shrink-0" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-accent-red">Reset Vault</p>
              <p className="text-2xs text-text-muted">
                Permanently delete all stored keys. This cannot be undone.
              </p>
            </div>
          </button>
        </div>
      </div>

      <ChangePasswordDrawer
        open={changeOpen}
        onClose={() => setChangeOpen(false)}
      />

      <ConfirmDialog
        open={resetOpen}
        onClose={() => {
          setResetConfirmText("");
          setResetOpen(false);
        }}
        onConfirm={() => {
          resetVault();
          setResetOpen(false);
        }}
        title="Reset Secure Vault"
        description={
          <>
            This will permanently delete all your stored SSH private keys. This
            action{" "}
            <strong className="text-text-primary">cannot be undone</strong>.
            Type <code className="text-accent-red font-mono">RESET</code> to
            confirm.
          </>
        }
        confirmLabel="Reset Vault"
        confirmDisabled={resetConfirmText !== "RESET"}
      >
        <div className="mb-4">
          <InputField
            id="vault-reset-confirm"
            label='Type "RESET" to confirm'
            hideLabel
            value={resetConfirmText}
            onChange={setResetConfirmText}
            placeholder="RESET"
            autoFocus
          />
        </div>
      </ConfirmDialog>
    </>
  );
}
