import { useState, useEffect, useId } from "react";
import { useWatch } from "react-hook-form";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "../stores/authStore";
import { useNamespaces } from "../hooks/useNamespaces";
import SectionedLayout from "@/components/settings/SectionedLayout";
import FormModal from "@/components/common/FormModal";
import ConfirmDialog from "../components/common/ConfirmDialog";
import BaseDialog from "../components/common/BaseDialog";
import CopyButton from "../components/common/CopyButton";
import { isSdkError } from "../api/errors";
import {
  editProfileSchema,
  type EditProfileFormValues,
  type CurrentProfileValues,
} from "./profile/editProfileSchema";
import {
  changePasswordSchema,
  type ChangePasswordFormValues,
} from "./profile/changePasswordSchema";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { isCloud, isCommunity } from "../env";
import {
  UserIcon,
  PencilSquareIcon,
  CheckIcon,
  UserCircleIcon,
  LockClosedIcon,
  TrashIcon,
  ExclamationTriangleIcon,
  CommandLineIcon,
  ArrowTopRightOnSquareIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import MfaEnableModal from "../components/mfa/MfaEnableModal";
import MfaDisableDialog from "../components/mfa/MfaDisableDialog";
import { isEnterpriseOrCloud } from "../env";
import { Button } from "@shellhub/design-system/primitives";
import PageLoader from "@/components/common/PageLoader";
import SettingsSection from "@/components/settings/SettingsSection";
import SettingsField from "@/components/settings/SettingsField";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import SettingsDangerCard from "@/components/settings/SettingsDangerCard";
import DialogHeader from "@/components/common/DialogHeader";

function DeleteAccountDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const deleteUser = useAuthStore((s) => s.deleteUser);
  const userId = useAuthStore((s) => s.userId);
  const { namespaces } = useNamespaces();
  const [error, setError] = useState("");

  const isNamespaceOwner = namespaces.some((ns) => ns.owner === userId);

  const handleDelete = async () => {
    setError("");
    try {
      await deleteUser();
    } catch (err) {
      if (isSdkError(err) && err.status === 403) {
        setError(
          "You cannot delete your account while you have active namespaces.",
        );
      } else {
        setError("Failed to delete account.");
      }
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
      confirmDisabled={isNamespaceOwner}
    >
      {(isNamespaceOwner || !!error) && (
        <div className="mb-4 space-y-2">
          {isNamespaceOwner && (
            <div className="p-3 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-start gap-2 text-accent-yellow">
              <ExclamationTriangleIcon
                className="w-4 h-4 shrink-0 mt-0.5"
                strokeWidth={2}
              />
              <span className="text-sm">
                Please delete all your owned namespaces before attempting to
                delete your account.
              </span>
            </div>
          )}
          {error && <p className="text-2xs text-accent-red">{error}</p>}
        </div>
      )}
    </ConfirmDialog>
  );
}

function DeleteAccountWarningDialog({
  open,
  onClose,
  isCommunity,
}: {
  open: boolean;
  onClose: () => void;
  isCommunity: boolean;
}) {
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
        icon={isCommunity ? <CommandLineIcon /> : <ShieldCheckIcon />}
        iconColor="neutral"
        title="Account deletion"
        description={
          isCommunity
            ? "On a Community instance, accounts are deleted from the CLI."
            : "On an Enterprise instance, accounts are deleted from the Admin Console."
        }
        titleId={accountDeletionTitleId}
        descriptionId={accountDeletionDescriptionId}
        onClose={onClose}
      />
      <div className="px-6 pb-6">
        <div className="space-y-4 text-sm text-text-muted">
          {isCommunity ? (
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
 * Edits the signed-in user's own name, username and email. Changing the email starts a
 * re-confirmation, so the account keeps the old address until the new one is verified.
 */
export function EditProfileModal({
  open,
  onClose,
  currentName,
  currentUsername,
  currentEmail,
  currentRecoveryEmail,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  currentUsername: string;
  currentEmail: string;
  currentRecoveryEmail: string;
}) {
  const updateProfile = useAuthStore((s) => s.updateProfile);

  const current: CurrentProfileValues = {
    name: currentName,
    username: currentUsername,
    email: currentEmail,
  };
  const schema = editProfileSchema(current);

  const form = useDrawerForm(open, schema, {
    name: currentName,
    username: currentUsername,
    email: currentEmail,
    recoveryEmail: currentRecoveryEmail,
  });
  const { control, setValue, trigger, setError, clearErrors, formState } = form;

  const onValid = async (values: EditProfileFormValues) => {
    clearErrors("root");

    const dirty = formState.dirtyFields;
    const data: {
      name?: string;
      username?: string;
      email?: string;
      recovery_email?: string;
    } = {};
    if (dirty.name) data.name = values.name;
    if (dirty.username) data.username = values.username;
    if (dirty.email) data.email = values.email;
    if (dirty.recoveryEmail) data.recovery_email = values.recoveryEmail;

    try {
      await updateProfile(data);
      onClose();
    } catch (err) {
      const status = isSdkError(err) ? err.status : undefined;
      const errorMessages: Record<number, string> = {
        400: "Some fields have invalid values. Review and try again.",
        409: "That username or email is already in use.",
      };
      const errorMessage =
        errorMessages[status ?? 0] ?? "Failed to update profile.";

      setError("root", { message: errorMessage });
    }
  };

  return (
    <FormModal
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<UserCircleIcon />}
      title="Edit profile"
      description="Your name, username, and the emails ShellHub signs you in with and writes to."
      submitLabel="Save"
      requireDirty
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <FormInputField
        name="name"
        control={control}
        id="profile-name"
        label="Name"
        placeholder="Your name"
        hint="1-64 characters"
        maxLength={64}
        onValueChange={() => clearErrors("root")}
      />
      <FormInputField
        name="username"
        control={control}
        id="profile-username"
        label="Username"
        labelAdornment={
          <span className="px-1.5 py-0.5 text-3xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
            Deprecated
          </span>
        }
        placeholder="username"
        hint="Lowercase letters, numbers, dots, underscores, @ and hyphens"
        maxLength={32}
        onValueChange={(v) => {
          setValue("username", v.toLowerCase(), {
            shouldDirty: true,
            shouldValidate: true,
          });
          clearErrors("root");
        }}
      />
      <FormInputField
        name="email"
        control={control}
        id="profile-email"
        label="Email"
        type="email"
        placeholder="you@example.com"
        onValueChange={() => {
          void trigger("recoveryEmail");
          clearErrors("root");
        }}
      />
      <FormInputField
        name="recoveryEmail"
        control={control}
        id="profile-recovery-email"
        label="Recovery Email"
        type="email"
        placeholder="recovery@example.com"
        hint="Optional. Used for account recovery if you lose access."
        onValueChange={() => clearErrors("root")}
      />
    </FormModal>
  );
}

function ChangePasswordModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const updatePw = useAuthStore((s) => s.updatePassword);
  const [success, setSuccess] = useState(false);

  const form = useDrawerForm(open, changePasswordSchema, {
    current: "",
    newPw: "",
    confirmPw: "",
  });
  const { control, setError, clearErrors } = form;

  const [watchedCurrent, watchedNewPw, watchedConfirmPw] = useWatch({
    control,
    name: ["current", "newPw", "confirmPw"],
  });
  const allFilled = !!watchedCurrent && !!watchedNewPw && !!watchedConfirmPw;

  useResetOnOpen(open, () => setSuccess(false));

  const onValid = async (values: ChangePasswordFormValues) => {
    clearErrors("root");
    try {
      await updatePw(values.current, values.newPw);
      setSuccess(true);
      setTimeout(onClose, 1200);
    } catch (err) {
      const errorMessage =
        isSdkError(err) && err.status === 403
          ? "Current password is incorrect."
          : "Failed to change password.";

      setError("root", { message: errorMessage });
    }
  };

  return (
    <FormModal
      size="sm"
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<LockClosedIcon />}
      title="Change password"
      description="Enter your current password, then the one to replace it."
      submitLabel="Change password"
      submitDisabled={!allFilled}
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <FormPasswordField
        name="current"
        control={control}
        id="change-pw-current"
        label="Current Password"
        autoComplete="current-password"
      />
      <FormPasswordField
        name="newPw"
        control={control}
        id="change-pw-new"
        label="New Password"
        autoComplete="new-password"
        hint="5-32 characters"
      />
      <FormPasswordField
        name="confirmPw"
        control={control}
        id="change-pw-confirm"
        label="Confirm New Password"
        autoComplete="new-password"
      />
      {success && (
        <p className="text-2xs text-accent-green">
          Password changed successfully.
        </p>
      )}
    </FormModal>
  );
}

/**
 * The profile section of the account: name, username, email and recovery email, and the modal
 * that edits them.
 */
export function AccountProfile() {
  const { name, username, email, recoveryEmail } = useAuthStore();
  const [editModalOpen, setEditModalOpen] = useState(false);

  if (!name && !username) {
    return <PageLoader label="Loading profile" padding="lg" />;
  }

  return (
    <>
      <SettingsSection
        title="Profile"
        description="How you appear in ShellHub and how it reaches you."
        action={
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setEditModalOpen(true)}
            icon={<PencilSquareIcon className="w-4 h-4" />}
          >
            Edit
          </Button>
        }
      >
        <SettingsField title="Name" description="Your display name.">
          <span className="text-sm font-mono text-text-secondary">{name}</span>
        </SettingsField>

        <SettingsField
          title="Username"
          description={
            <>
              The old login identifier; sign in with your email instead.
              <span className="ml-2 px-1.5 py-0.5 text-3xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                Deprecated
              </span>
            </>
          }
        >
          <span className="text-sm font-mono text-text-secondary">
            {username}
          </span>
        </SettingsField>

        <SettingsField
          title="Email"
          description="How you sign in and where account mail goes."
        >
          <span className="text-sm font-mono text-text-secondary">{email}</span>
        </SettingsField>

        <SettingsField
          title="Recovery email"
          description="Where account recovery goes if you lose access."
        >
          <span className="text-sm font-mono text-text-secondary">
            {recoveryEmail || (
              <span className="text-text-muted italic font-sans">Not set</span>
            )}
          </span>
        </SettingsField>
      </SettingsSection>
      <EditProfileModal
        open={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        currentName={name ?? ""}
        currentUsername={username ?? ""}
        currentEmail={email ?? ""}
        currentRecoveryEmail={recoveryEmail ?? ""}
      />
    </>
  );
}

/**
 * The security section of the account: the password and multi-factor authentication, or a note
 * that the identity provider manages both for an SSO user.
 */
export function AccountSecurity() {
  const { origin, mfaEnabled, recoveryEmail, fetchUser } = useAuthStore();
  const isSsoUser = origin === "saml";
  const [pwModalOpen, setPwModalOpen] = useState(false);
  const [mfaEnableOpen, setMfaEnableOpen] = useState(false);
  const [mfaDisableOpen, setMfaDisableOpen] = useState(false);

  return (
    <>
      <SettingsSection
        title="Security"
        description="How you prove it's you when you sign in."
      >
        {isSsoUser ? (
          <SettingsField
            title="Managed by your identity provider"
            description="You sign in through SSO, so your password and multi-factor authentication are set with your identity provider, not in ShellHub."
          >
            <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-primary/10 text-primary border border-primary/20">
              SSO
            </span>
          </SettingsField>
        ) : (
          <>
            <SettingsField
              title="Password"
              description="The password you sign in with."
            >
              <Button
                size="sm"
                variant="secondary"
                onClick={() => setPwModalOpen(true)}
              >
                Change Password
              </Button>
            </SettingsField>

            <SettingsSwitchCard
              icon={<ShieldCheckIcon />}
              title="Multi-factor authentication"
              description={
                isEnterpriseOrCloud()
                  ? "A code from an authenticator app on every sign-in. Recovery codes are shown once, during setup."
                  : "A code from an authenticator app on every sign-in. Available on paid editions."
              }
              control={
                !isEnterpriseOrCloud() ? (
                  <Button
                    as="a"
                    size="sm"
                    variant="secondary"
                    href="https://www.shellhub.io/pricing"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Upgrade
                  </Button>
                ) : mfaEnabled ? (
                  <div className="flex items-center gap-3">
                    <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-green/10 text-accent-green border border-accent-green/20">
                      Enabled
                    </span>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => setMfaDisableOpen(true)}
                    >
                      Disable
                    </Button>
                  </div>
                ) : (
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => setMfaEnableOpen(true)}
                  >
                    Enable MFA
                  </Button>
                )
              }
            />
          </>
        )}
      </SettingsSection>
      <ChangePasswordModal
        open={pwModalOpen}
        onClose={() => setPwModalOpen(false)}
      />
      <MfaEnableModal
        open={mfaEnableOpen}
        onClose={() => setMfaEnableOpen(false)}
        onSuccess={() => {
          setMfaEnableOpen(false);
          void fetchUser();
        }}
        currentRecoveryEmail={recoveryEmail ?? null}
      />
      <MfaDisableDialog
        open={mfaDisableOpen}
        onClose={() => setMfaDisableOpen(false)}
        onSuccess={() => {
          setMfaDisableOpen(false);
          void fetchUser();
        }}
      />
    </>
  );
}

/**
 * The irreversible section of the account: deleting it, which only the cloud does from here.
 */
export function AccountDangerZone() {
  const isCloudEdition = isCloud();
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
            isCloudEdition
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
      {isCloudEdition ? (
        <DeleteAccountDialog
          key={String(deleteDialogOpen)}
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
        />
      ) : (
        <DeleteAccountWarningDialog
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
          isCommunity={isCommunity()}
        />
      )}
    </>
  );
}

const ACCOUNT_SECTIONS = [
  { to: "profile", label: "Profile", icon: UserIcon },
  { to: "security", label: "Security", icon: ShieldCheckIcon },
  { to: "danger-zone", label: "Danger zone", icon: ExclamationTriangleIcon },
];

/**
 * The user's own account, one section per URL under /account: who they are, how they sign in, and
 * deleting the account.
 */
export default function Profile() {
  const fetchUser = useAuthStore((s) => s.fetchUser);

  useEffect(() => {
    void fetchUser();
  }, [fetchUser]);

  return (
    <SectionedLayout
      base="/account"
      icon={<UserIcon className="w-6 h-6" />}
      title="Account"
      description="Your profile and how you sign in"
      sections={ACCOUNT_SECTIONS}
    />
  );
}
