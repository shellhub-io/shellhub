import { useState, useId } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaceRequests";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import BaseDialog from "@/components/common/BaseDialog";
import CopyButton from "@/components/common/CopyButton";
import InputField from "@/components/common/fields/InputField";
import PasswordField from "@/components/common/fields/PasswordField";
import { isSdkError, lockoutMessage } from "@/api/errors";

import { features } from "@/edition";
import {
  TrashIcon,
  ExclamationTriangleIcon,
  CommandLineIcon,
  ArrowTopRightOnSquareIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsDangerCard from "@/components/settings/SettingsDangerCard";
import DialogHeader from "@/components/common/DialogHeader";

const OWNS_NAMESPACE_MESSAGE = "user still has a owned namespace";

function deleteAccountErrorMessage(err: unknown, mfaEnabled: boolean): string {
  if (isSdkError(err) && err.status === 403) {
    if (err.message === OWNS_NAMESPACE_MESSAGE)
      return "You cannot delete your account while you own namespaces.";

    return mfaEnabled ? "The code is incorrect." : "The password is incorrect.";
  }

  return lockoutMessage(err) ?? "Failed to delete account.";
}

function DeleteAccountDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const deleteUser = useAuthStore((s) => s.deleteUser);
  const userId = useAuthStore((s) => s.userId);
  const mfaEnabled = useAuthStore((s) => s.mfaEnabled);
  const isSsoUser = useAuthStore((s) => s.origin === "saml");
  const { namespaces } = useNamespaces();
  const [factor, setFactor] = useState("");
  const [error, setError] = useState("");

  const isNamespaceOwner = namespaces.some((ns) => ns.owner === userId);

  const handleDelete = async () => {
    setError("");
    try {
      await deleteUser(mfaEnabled ? { code: factor } : { password: factor });
    } catch (err) {
      setFactor("");
      setError(deleteAccountErrorMessage(err, mfaEnabled));
    }
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleDelete}
      icon={<TrashIcon />}
      title="Delete account"
      description={
        isNamespaceOwner
          ? "You can't delete your account while you own namespaces."
          : "Your account is deleted. This can't be undone."
      }
      confirmLabel="Delete account"
      confirmDisabled={isNamespaceOwner || isSsoUser || factor === ""}
      errorMessage={error}
    >
      {isNamespaceOwner ? (
        <div className="p-3 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-start gap-2 text-accent-yellow">
          <ExclamationTriangleIcon
            className="w-4 h-4 shrink-0 mt-0.5"
            strokeWidth={2}
          />
          <span className="text-sm">
            Please delete all your owned namespaces before attempting to delete
            your account.
          </span>
        </div>
      ) : isSsoUser ? (
        <p className="text-sm text-text-muted">
          Your account signs in through your identity provider, so it has no
          password to confirm the deletion with. Ask an administrator to delete
          it.
        </p>
      ) : mfaEnabled ? (
        <InputField
          id="delete-account-code"
          label="Verification code"
          value={factor}
          onChange={setFactor}
          hint="Enter the code from your authenticator app."
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
        />
      ) : (
        <PasswordField
          id="delete-account-password"
          label="Password"
          value={factor}
          onChange={setFactor}
          hint="Enter your password to confirm."
          autoComplete="current-password"
        />
      )}
    </ConfirmDialog>
  );
}

function DeleteAccountWarningDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const viaCli = !features().paid;
  const username = useAuthStore((s) => s.username);
  const userId = useAuthStore((s) => s.userId);
  const { namespaces } = useNamespaces();

  const isNamespaceOwner = namespaces.some((ns) => ns.owner === userId);
  const deleteCommand = `./bin/cli user delete ${username ?? ""}`;
  const accountDeletionTitleId = useId();
  const accountDeletionDescriptionId = useId();

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-labelledby={accountDeletionTitleId}
      aria-describedby={accountDeletionDescriptionId}
    >
      <DialogHeader
        icon={viaCli ? <CommandLineIcon /> : <ShieldCheckIcon />}
        iconColor="neutral"
        title="Account deletion"
        description={
          viaCli
            ? "On a Community instance, accounts are deleted from the CLI."
            : "On an Enterprise instance, accounts are deleted from the Admin Console."
        }
        titleId={accountDeletionTitleId}
        descriptionId={accountDeletionDescriptionId}
        onClose={onClose}
      />
      <div className="px-6 pb-6">
        <div className="space-y-4 text-sm text-text-muted">
          {viaCli ? (
            <>
              <p>
                In Community instances, user accounts can only be deleted via
                the CLI. For detailed instructions, refer to our{" "}
                <a
                  href="https://docs.shellhub.io/self-hosted/administration#delete-a-user"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline inline-flex items-center gap-1"
                  data-test="docs-link"
                >
                  administration documentation
                  <ArrowTopRightOnSquareIcon className="w-3.5 h-3.5" />
                </a>
                .
              </p>
              <div>
                <p className="text-2xs font-medium text-text-secondary mb-1.5">
                  Run this command to delete your account:
                </p>
                <div className="flex items-center gap-2 bg-hover-medium border border-border rounded-lg px-3 py-2">
                  <span className="flex-1 truncate font-mono text-2xs text-text-primary">
                    {deleteCommand}
                  </span>
                  <CopyButton text={deleteCommand} size="sm" />
                </div>
              </div>
              {isNamespaceOwner && (
                <div className="p-3 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-start gap-2 text-accent-yellow">
                  <ExclamationTriangleIcon
                    className="w-4 h-4 shrink-0 mt-0.5"
                    strokeWidth={2}
                  />
                  <span className="text-2xs">
                    <strong>Namespace owner:</strong> You own one or more
                    namespaces. You must delete all owned namespaces before
                    deleting your account.
                  </span>
                </div>
              )}
            </>
          ) : (
            <p>
              In Enterprise instances, user accounts can only be deleted via the
              Admin Console. Please access your{" "}
              <a
                href="/admin/users"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary hover:underline"
              >
                Admin Console
              </a>{" "}
              or contact your system administrator for assistance.
            </p>
          )}
        </div>

        <div className="flex justify-end mt-6">
          <Button variant="ghost" onClick={onClose} data-test="close-btn">
            Close
          </Button>
        </div>
      </div>
    </BaseDialog>
  );
}

/**
 * The irreversible section of the account: deleting it, which only the cloud does from here.
 */
export default function AccountDangerZone() {
  const { selfService } = features();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  return (
    <>
      <SettingsSection
        title="Danger zone"
        description="Actions on your account that can't be undone."
      >
        <SettingsDangerCard
          title="Delete account"
          description={
            selfService
              ? "Removes your account and everything tied to it for good."
              : "Account deletion needs the CLI or the Admin Console."
          }
          action={
            <Button
              size="sm"
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
              data-test="delete-account-btn"
            >
              Delete account
            </Button>
          }
        />
      </SettingsSection>
      {selfService ? (
        <DeleteAccountDialog
          key={String(deleteDialogOpen)}
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
        />
      ) : (
        <DeleteAccountWarningDialog
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
        />
      )}
    </>
  );
}
