import { useState, useEffect, useRef, FormEvent } from "react";
import { Card, Spinner } from "@shellhub/design-system/primitives";
import {
  KeyIcon,
  LockClosedIcon,
  ExclamationTriangleIcon,
  ExclamationCircleIcon,
  ChevronDownIcon,
  ServerStackIcon,
} from "@heroicons/react/24/outline";
import { useVaultStore } from "@/stores/vaultStore";
import { isVaultServerEnabled } from "@/utils/vault-backend-factory";
import VaultSyncDialog from "@/components/vault/VaultSyncDialog";
import { ALLOWED_TIMEOUT_MINUTES } from "@/types/vault";
import type { AllowedTimeoutMinutes } from "@/types/vault";
import { useClickOutside } from "@/hooks/useClickOutside";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import CheckboxField from "@/components/common/fields/CheckboxField";
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
            {loading && <Spinner size="sm" tone="onPrimary" />}
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

const TIMEOUT_LABELS: Record<AllowedTimeoutMinutes, string> = {
  0: "Never",
  5: "5 minutes",
  15: "15 minutes",
  30: "30 minutes",
  60: "60 minutes",
};

function AutoLockTimeoutSelect({
  value,
  onChange,
}: {
  value: number;
  onChange: (minutes: AllowedTimeoutMinutes) => void;
}) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  useClickOutside(containerRef, () => setOpen(false));

  const currentLabel =
    TIMEOUT_LABELS[value as AllowedTimeoutMinutes] ?? `${value} minutes`;

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        aria-label="Auto-lock timeout"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm text-text-primary bg-card border border-border rounded-md hover:border-border-light transition-colors"
      >
        {currentLabel}
        <ChevronDownIcon
          className={`w-3.5 h-3.5 text-text-muted transition-transform duration-150 ${open ? "rotate-180" : ""}`}
          strokeWidth={2.5}
        />
      </button>

      {open && (
        <ul
          role="listbox"
          aria-label="Auto-lock timeout options"
          className="absolute right-0 top-full mt-1 w-36 bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down"
        >
          {ALLOWED_TIMEOUT_MINUTES.map((minutes) => (
            <li
              key={minutes}
              role="option"
              aria-selected={value === minutes}
              onClick={() => {
                onChange(minutes);
                setOpen(false);
              }}
              className={`px-3 py-2 text-sm cursor-pointer transition-colors ${
                value === minutes
                  ? "text-primary bg-primary/10"
                  : "text-text-secondary hover:text-text-primary hover:bg-hover-medium"
              }`}
            >
              {TIMEOUT_LABELS[minutes]}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default function VaultSettingsSection() {
  const status = useVaultStore((s) => s.status);
  const lock = useVaultStore((s) => s.lock);
  const resetVault = useVaultStore((s) => s.resetVault);
  const autoLockTimeoutMinutes = useVaultStore((s) => s.autoLockTimeoutMinutes);
  const lockOnHidden = useVaultStore((s) => s.lockOnHidden);
  const updateAutoLockSettings = useVaultStore((s) => s.updateAutoLockSettings);
  const storageMode = useVaultStore((s) => s.storageMode);
  const [changeOpen, setChangeOpen] = useState(false);
  const [resetOpen, setResetOpen] = useState(false);
  const [resetConfirmText, setResetConfirmText] = useState("");
  const [syncOpen, setSyncOpen] = useState(false);

  if (status !== "unlocked") return null;

  return (
    <>
      <div className="mt-8 animate-fade-in">
        <h3 className="text-xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
          Vault Settings
        </h3>
        <Card className="divide-y divide-border">
          <button
            type="button"
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

          <div className="flex items-center gap-3 w-full px-4 py-3.5">
            <LockClosedIcon className="w-4 h-4 text-text-muted shrink-0" />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-text-primary">
                Auto-lock Timeout
              </p>
              <p className="text-2xs text-text-muted">
                Automatically lock the vault after this period of inactivity.
              </p>
            </div>
            <AutoLockTimeoutSelect
              value={autoLockTimeoutMinutes}
              onChange={(minutes) =>
                void updateAutoLockSettings({ autoLockTimeoutMinutes: minutes })
              }
            />
          </div>

          <div className="flex items-center gap-3 w-full px-4 py-3.5">
            <div className="min-w-0 flex-1">
              <CheckboxField
                id="vault-lock-on-hidden"
                label="Lock when hidden"
                description="Locks the vault about a minute after you switch away or minimize."
                checked={lockOnHidden}
                onChange={(checked) =>
                  void updateAutoLockSettings({ lockOnHidden: checked })
                }
              />
            </div>
          </div>

          {isVaultServerEnabled() && (
            <button
              type="button"
              onClick={() => setSyncOpen(true)}
              className="flex items-center gap-3 w-full px-4 py-3.5 text-left hover:bg-hover-subtle transition-colors"
            >
              <ServerStackIcon className="w-4 h-4 text-text-muted shrink-0" />
              <div className="min-w-0">
                <p className="text-sm font-medium text-text-primary">Storage</p>
                <p className="text-2xs text-text-muted">
                  {storageMode === "server"
                    ? "Synced with the ShellHub server. Click to move it to this device."
                    : "Stored in this browser only. Click to sync it to the ShellHub server."}
                </p>
              </div>
            </button>
          )}

          <button
            type="button"
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
            type="button"
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
        </Card>
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
          void resetVault();
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

      <VaultSyncDialog
        open={syncOpen}
        onClose={() => setSyncOpen(false)}
        direction={storageMode === "server" ? "to-local" : "to-server"}
      />
    </>
  );
}
