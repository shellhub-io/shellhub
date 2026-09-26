import { useState, FormEvent, useEffect, useId } from "react";
import {
  ShieldCheckIcon,
  ExclamationTriangleIcon,
  ComputerDesktopIcon,
  ServerStackIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useVaultStore } from "@/stores/vaultStore";
import { loadLegacyKeysFromStorage } from "@/utils/vault-backend-local";
import {
  isVaultServerEnabled,
  type VaultStorageMode,
} from "@/utils/vault-backend-factory";
import BaseDialog from "@/components/common/BaseDialog";
import PasswordField from "@/components/common/fields/PasswordField";
import { Button } from "@shellhub/design-system/primitives";
import DialogHeader from "@/components/common/DialogHeader";

interface Props {
  open: boolean;
  onClose: () => void;
}

interface FormProps extends Props {
  instanceId: string;
}

const STORAGE_OPTIONS: {
  mode: VaultStorageMode;
  icon: typeof ServerStackIcon;
  title: string;
  description: string;
}[] = [
  {
    mode: "server",
    icon: ServerStackIcon,
    title: "Sync to the ShellHub server",
    description:
      "Use your keys on any machine you sign in to. Stored encrypted — the server never sees them.",
  },
  {
    mode: "local",
    icon: ComputerDesktopIcon,
    title: "This device only",
    description:
      "Keys stay in this browser. Clearing its data deletes them, and other machines can't reach them.",
  },
];

function SetupForm({ open, onClose, instanceId }: FormProps) {
  const loading = useVaultStore((s) => s.loading);
  const error = useVaultStore((s) => s.error);
  const initialize = useVaultStore((s) => s.initialize);
  const clearError = useVaultStore((s) => s.clearError);
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [mode, setMode] = useState<VaultStorageMode>("server");

  const serverEnabled = isVaultServerEnabled();
  const titleId = `vault-setup-title-${instanceId}`;

  useEffect(() => {
    if (open) clearError();
  }, [open, clearError]);

  const legacyCount = open ? loadLegacyKeysFromStorage().length : 0;

  const passwordTooShort = password.length > 0 && password.length < 8;
  const passwordsMismatch = confirm.length > 0 && password !== confirm;
  const canSubmit = password.length >= 8 && password === confirm && !loading;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    await initialize(password, serverEnabled ? mode : "local");
    if (useVaultStore.getState().status === "unlocked") {
      onClose();
    }
  };

  return (
    <div>
      <DialogHeader
        icon={<ShieldCheckIcon />}
        title="Set up secure vault"
        description="Encrypt your private keys with a master password."
        titleId={titleId}
        descriptionId={`${titleId}-description`}
        onClose={onClose}
      />
      <div className="px-6 pb-6">
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
          {serverEnabled && (
            <fieldset className="space-y-2">
              <legend className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
                Where to store it
              </legend>
              {STORAGE_OPTIONS.map((option) => {
                const selected = mode === option.mode;
                const Icon = option.icon;
                return (
                  <label
                    key={option.mode}
                    className={cn(
                      "flex items-start gap-3 px-3.5 py-3 rounded-lg border cursor-pointer transition-colors",
                      selected
                        ? "border-primary bg-primary/[0.06]"
                        : "border-border hover:border-border-light hover:bg-hover-subtle",
                    )}
                  >
                    <input
                      type="radio"
                      name={`${instanceId}-storage`}
                      value={option.mode}
                      checked={selected}
                      onChange={() => setMode(option.mode)}
                      className="sr-only"
                    />
                    <Icon
                      className={cn(
                        "w-5 h-5 shrink-0 mt-0.5",
                        selected ? "text-primary" : "text-text-muted",
                      )}
                      strokeWidth={2}
                    />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-1.5">
                        <span className="text-sm font-medium text-text-primary">
                          {option.title}
                        </span>
                        {selected && (
                          <CheckCircleIcon
                            className="w-4 h-4 text-primary shrink-0"
                            strokeWidth={2}
                          />
                        )}
                      </div>
                      <p className="text-2xs text-text-muted mt-0.5">
                        {option.description}
                      </p>
                    </div>
                  </label>
                );
              })}
            </fieldset>
          )}

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
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={!canSubmit} loading={loading}>
              Create Vault
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

/**
 * Creates a vault. The passphrase entered here is the only way into it afterwards: it is never
 * sent anywhere and cannot be reset, only discarded along with the keys.
 */
export default function VaultSetupDialog({ open, onClose }: Props) {
  const instanceId = useId();
  const titleId = `vault-setup-title-${instanceId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={`${titleId}-description`}
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
