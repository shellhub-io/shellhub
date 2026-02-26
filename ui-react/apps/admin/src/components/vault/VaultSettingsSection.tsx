import { useState, FormEvent } from "react";
import {
  KeyIcon,
  LockClosedIcon,
  ExclamationTriangleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import Drawer from "@/components/common/Drawer";
import { INPUT } from "@/utils/styles";

function ChangePasswordDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const { loading, error, changeMasterPassword } = useVaultStore();
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

  const newTooShort = newPassword.length > 0 && newPassword.length < 8;
  const mismatch = confirmPassword.length > 0 && newPassword !== confirmPassword;
  const canSubmit =
    currentPassword.length > 0 &&
    newPassword.length >= 8 &&
    newPassword === confirmPassword &&
    !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
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
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {loading && (
              <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            )}
            Update Password
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label
            htmlFor="vault-current-password"
            className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
          >
            Current Password
          </label>
          <input
            id="vault-current-password"
            type="password"
            autoComplete="off"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            placeholder="Enter current master password"
            autoFocus={open}
            className={INPUT}
          />
        </div>

        <div className="h-px bg-border" />

        <div>
          <label
            htmlFor="vault-new-password"
            className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
          >
            New Password
          </label>
          <input
            id="vault-new-password"
            type="password"
            autoComplete="off"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Minimum 8 characters"
            className={INPUT}
          />
          {newTooShort && (
            <p className="text-2xs text-accent-red mt-1.5">
              Password must be at least 8 characters
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="vault-confirm-new-password"
            className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1.5"
          >
            Confirm New Password
          </label>
          <input
            id="vault-confirm-new-password"
            type="password"
            autoComplete="off"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Re-enter new password"
            className={INPUT}
          />
          {mismatch && (
            <p className="text-2xs text-accent-red mt-1.5">
              Passwords do not match
            </p>
          )}
        </div>

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
  const { status, lock, resetVault } = useVaultStore();
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
              <p className="text-sm font-medium text-accent-red">
                Reset Vault
              </p>
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
        onClose={() => setResetOpen(false)}
        onConfirm={async () => {
          await resetVault();
          setResetOpen(false);
        }}
        title="Reset Secure Vault"
        description={
          <>
            This will permanently delete all your stored SSH private keys. This
            action <strong className="text-text-primary">cannot be undone</strong>.
            Type <code className="text-accent-red font-mono">RESET</code> to confirm.
          </>
        }
        confirmLabel="Reset Vault"
        confirmDisabled={resetConfirmText !== "RESET"}
      >
        <div className="mb-4">
          <input
            type="text"
            value={resetConfirmText}
            onChange={(e) => setResetConfirmText(e.target.value)}
            placeholder='Type "RESET" to confirm'
            className={INPUT}
          />
        </div>
      </ConfirmDialog>
    </>
  );
}
