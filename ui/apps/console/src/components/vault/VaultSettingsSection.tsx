import { useState, useEffect, FormEvent } from "react";
import { Button, Dropdown, Toggle } from "@shellhub/design-system/primitives";
import {
  KeyIcon,
  LockClosedIcon,
  ExclamationCircleIcon,
  ChevronDownIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useVaultStore } from "@/stores/vaultStore";
import { isVaultServerEnabled } from "@/utils/vault-backend-factory";
import VaultSyncDialog from "@/components/vault/VaultSyncDialog";
import { ALLOWED_TIMEOUT_MINUTES } from "@/types/vault";
import type { AllowedTimeoutMinutes } from "@/types/vault";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import Modal from "@/components/common/Modal";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import SettingsDangerCard from "@/components/settings/SettingsDangerCard";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import ObjectName from "@/components/common/ObjectName";
function ChangePasswordModal({
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
    <Modal
      size="sm"
      open={open}
      onClose={onClose}
      icon={<KeyIcon />}
      title="Change master password"
      description="The vault re-encrypts your keys with the new password. The keys themselves don't change."
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
    </Modal>
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
  const currentLabel =
    TIMEOUT_LABELS[value as AllowedTimeoutMinutes] ?? `${value} minutes`;

  return (
    <Dropdown placement="bottom-end" open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label="Auto-lock timeout"
          className="flex items-center gap-1.5 px-2.5 py-1.5 text-sm text-text-primary bg-card border border-border rounded-md hover:border-border-light transition-colors"
        >
          {currentLabel}
          <ChevronDownIcon
            className={cn(
              "w-3.5 h-3.5 text-text-muted transition-transform",
              open && "rotate-180",
            )}
            strokeWidth={2.5}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel aria-label="Auto-lock timeout options" className="w-36">
        {ALLOWED_TIMEOUT_MINUTES.map((minutes) => (
          <Dropdown.Item
            key={minutes}
            label={TIMEOUT_LABELS[minutes]}
            onSelect={() => onChange(minutes)}
            className={cn(
              "px-3 py-2 text-sm",
              value === minutes && "text-primary bg-primary/10",
            )}
          >
            {TIMEOUT_LABELS[minutes]}
          </Dropdown.Item>
        ))}
      </Dropdown.Panel>
    </Dropdown>
  );
}

/**
 * The vault's settings on the Secure Vault page: lock timing, where the vault is stored, and the
 * reset. Resetting destroys the keys, and the passphrase cannot be recovered, so it is guarded
 * here.
 */
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
      <div className="mt-10 space-y-12 animate-fade-in">
        <SettingsSection
          title="Vault settings"
          description="How the vault locks, where it is kept, and its master password."
        >
          <SettingsField
            title="Master password"
            description="Re-encrypts every key in the vault with a new password."
          >
            <Button
              size="sm"
              variant="secondary"
              aria-label="Change master password"
              onClick={() => setChangeOpen(true)}
            >
              Change
            </Button>
          </SettingsField>

          <SettingsField
            title="Auto-lock timeout"
            description="Locks the vault after this long without use."
          >
            <AutoLockTimeoutSelect
              value={autoLockTimeoutMinutes}
              onChange={(minutes) =>
                void updateAutoLockSettings({ autoLockTimeoutMinutes: minutes })
              }
            />
          </SettingsField>

          <SettingsSwitchCard
            icon={<LockClosedIcon />}
            title="Lock when hidden"
            description="Locks the vault about a minute after you switch away or minimize the window."
            control={
              <Toggle
                aria-label="Lock when hidden"
                enabled={lockOnHidden}
                onChange={(checked) =>
                  void updateAutoLockSettings({ lockOnHidden: checked })
                }
              />
            }
          />

          {isVaultServerEnabled() && (
            <SettingsField title="Storage" description={storageDescription}>
              <Button
                size="sm"
                variant="secondary"
                aria-label="Change vault storage location"
                onClick={() => setSyncOpen(true)}
              >
                {storageMode === "server" ? "Move" : "Sync"}
              </Button>
            </SettingsField>
          )}

          <SettingsField
            title="Lock now"
            description="Clears the decrypted keys from memory."
          >
            <Button
              size="sm"
              variant="secondary"
              aria-label="Lock vault"
              onClick={lock}
            >
              Lock
            </Button>
          </SettingsField>
        </SettingsSection>

        <SettingsSection
          title="Danger zone"
          description="Actions on the vault that can't be undone."
        >
          <SettingsDangerCard
            title="Reset vault"
            description="Deletes every key stored in the vault for good."
            action={
              <Button
                size="sm"
                variant="destructive"
                aria-label="Reset vault"
                onClick={() => {
                  setResetConfirmText("");
                  setResetOpen(true);
                }}
              >
                Reset vault
              </Button>
            }
          />
        </SettingsSection>
      </div>

      <ChangePasswordModal
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
        icon={<TrashIcon />}
        title="Reset secure vault"
        description={
          <>
            Every private key stored in the vault is deleted, and this can't be
            undone. Type <ObjectName>RESET</ObjectName> to confirm.
          </>
        }
        confirmLabel="Reset vault"
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
