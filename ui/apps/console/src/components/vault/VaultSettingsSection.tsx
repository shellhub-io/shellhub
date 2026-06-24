import { useState, useEffect, useRef, FormEvent } from "react";
import { Button } from "@shellhub/design-system/primitives";
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
import SettingsCard from "@/components/common/SettingsCard";
import SettingsRow from "@/components/common/SettingsRow";
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
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={!canSubmit}
            loading={loading}
          >
            Update Password
          </Button>
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

  const storageDescription =
    storageMode === "server"
      ? "Synced with the ShellHub server. Click to move it to this device."
      : "Stored in this browser only. Click to sync it to the ShellHub server.";

  return (
    <>
      <div className="mt-8 space-y-4 animate-fade-in">
        <SettingsCard title="Vault Settings">
          <SettingsRow
            icon={<KeyIcon className="w-4 h-4" />}
            title="Change Master Password"
            description="Re-encrypt all keys with a new password."
          >
            <Button
              size="sm"
              variant="secondary"
              aria-label="Change master password"
              onClick={() => setChangeOpen(true)}
            >
              Change
            </Button>
          </SettingsRow>

          <SettingsRow
            icon={<LockClosedIcon className="w-4 h-4" />}
            title="Auto-lock Timeout"
            description="Automatically lock the vault after this period of inactivity."
          >
            <AutoLockTimeoutSelect
              value={autoLockTimeoutMinutes}
              onChange={(minutes) =>
                void updateAutoLockSettings({ autoLockTimeoutMinutes: minutes })
              }
            />
          </SettingsRow>

          <SettingsRow
            icon={<LockClosedIcon className="w-4 h-4" />}
            title="Lock when hidden"
            description="Locks the vault about a minute after you switch away or minimize."
          >
            <CheckboxField
              id="vault-lock-on-hidden"
              label="Lock when hidden"
              hideLabel
              aria-label="Lock when hidden"
              checked={lockOnHidden}
              onChange={(checked) =>
                void updateAutoLockSettings({ lockOnHidden: checked })
              }
            />
          </SettingsRow>

          {isVaultServerEnabled() && (
            <SettingsRow
              icon={<ServerStackIcon className="w-4 h-4" />}
              title="Storage"
              description={storageDescription}
            >
              <Button
                size="sm"
                variant="secondary"
                aria-label="Change vault storage location"
                onClick={() => setSyncOpen(true)}
              >
                {storageMode === "server" ? "Move" : "Sync"}
              </Button>
            </SettingsRow>
          )}

          <SettingsRow
            icon={<LockClosedIcon className="w-4 h-4" />}
            title="Lock Vault"
            description="Clear decrypted keys from memory."
          >
            <Button
              size="sm"
              variant="secondary"
              aria-label="Lock vault"
              onClick={lock}
            >
              Lock
            </Button>
          </SettingsRow>
        </SettingsCard>

        <SettingsCard title="Danger Zone" danger>
          <SettingsRow
            icon={<ExclamationTriangleIcon className="w-4 h-4" />}
            title="Reset Vault"
            description="Permanently delete all stored keys. This cannot be undone."
          >
            <Button
              size="sm"
              variant="dangerSoft"
              aria-label="Reset vault"
              onClick={() => {
                setResetConfirmText("");
                setResetOpen(true);
              }}
            >
              Reset
            </Button>
          </SettingsRow>
        </SettingsCard>
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
