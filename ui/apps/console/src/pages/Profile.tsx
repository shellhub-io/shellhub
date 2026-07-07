import { useState, useEffect, useLayoutEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useAuthStore } from "../stores/authStore";
import { useNamespaces } from "../hooks/useNamespaces";
import PageHeader from "../components/common/PageHeader";
import Drawer from "../components/common/Drawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import BaseDialog from "../components/common/BaseDialog";
import CopyButton from "../components/common/CopyButton";
import { isSdkError } from "../api/errors";
import { rhfEditProfileResolver, type EditProfileFormValues } from "./profile/editProfileResolver";
import { rhfChangePasswordResolver, type ChangePasswordFormValues } from "./profile/changePasswordResolver";
import FormPasswordField from "@/components/common/fields/rhf/FormPasswordField";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import { getConfig } from "../env";
import {
  UserIcon,
  PencilSquareIcon,
  CheckIcon,
  UserCircleIcon,
  EnvelopeIcon,
  LockClosedIcon,
  TrashIcon,
  AtSymbolIcon,
  ExclamationTriangleIcon,
  CommandLineIcon,
  ArrowTopRightOnSquareIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import MfaEnableDrawer from "../components/mfa/MfaEnableDrawer";
import MfaDisableDialog from "../components/mfa/MfaDisableDialog";
import { hasMfaSupport } from "../utils/features";
import { Button } from "@shellhub/design-system/primitives";
import PageLoader from "@/components/common/PageLoader";
import SettingsCard from "@/components/common/SettingsCard";
import SettingsRow from "@/components/common/SettingsRow";

/* ─── Delete Account Dialog (Cloud) ─── */

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
      title="Confirm Account Deletion"
      description={
        isNamespaceOwner
          ? "You cannot delete your account while you have active namespaces."
          : "Are you sure you want to delete your account? This action cannot be undone."
      }
      confirmLabel="Delete Account"
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

/* ─── Delete Account Warning Dialog (Community / Enterprise) ─── */

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

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-label="Account Deletion"
    >
      <div className="p-6">
        <div className="flex items-start gap-3 mb-5">
          <span className="w-9 h-9 rounded-lg bg-hover-medium border border-border flex items-center justify-center shrink-0">
            {isCommunity ? (
              <CommandLineIcon className="w-5 h-5 text-text-muted" />
            ) : (
              <ShieldCheckIcon className="w-5 h-5 text-text-muted" />
            )}
          </span>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              Account Deletion
            </h2>
            <p className="text-2xs text-text-muted mt-0.5">
              {isCommunity ? "CLI Required" : "Admin Console Required"}
            </p>
          </div>
        </div>

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

/* ─── Edit Profile Drawer ─── */

export function EditProfileDrawer({
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

  const { control, handleSubmit, reset, setValue, trigger, setError, clearErrors, formState } =
    useForm<EditProfileFormValues>({
      resolver: rhfEditProfileResolver,
      mode: "onChange",
      defaultValues: {
        name: currentName,
        username: currentUsername,
        email: currentEmail,
        recoveryEmail: currentRecoveryEmail,
      },
    });

  const { isDirty, isValid, isSubmitting } = formState;

  useLayoutEffect(() => {
    reset({
      name: currentName,
      username: currentUsername,
      email: currentEmail,
      recoveryEmail: currentRecoveryEmail,
    });
  }, [open, currentName, currentUsername, currentEmail, currentRecoveryEmail, reset]);

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
      if (status === 409)
        setError("root", { message: "That username or email is already in use." });
      else if (status === 400)
        setError("root", { message: "Some fields have invalid values. Review and try again." });
      else setError("root", { message: "Failed to update profile." });
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Profile"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit(onValid)()}
            disabled={!isDirty || !isValid || isSubmitting}
            loading={isSubmitting}
            icon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Save
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(onValid)(e)} className="space-y-5">
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
            setValue("username", v.toLowerCase(), { shouldDirty: true, shouldValidate: true });
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
        {formState.errors.root && (
          <p className="text-2xs text-accent-red">{formState.errors.root.message}</p>
        )}
      </form>
    </Drawer>
  );
}

/* ─── Change Password Drawer ─── */

function ChangePasswordDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const updatePw = useAuthStore((s) => s.updatePassword);
  const [success, setSuccess] = useState(false);

  const { control, handleSubmit, reset, setError, formState } =
    useForm<ChangePasswordFormValues>({
      resolver: rhfChangePasswordResolver,
      mode: "onChange",
      defaultValues: { current: "", newPw: "", confirmPw: "" },
    });

  const { isValid, isSubmitting } = formState;
  const [watchedCurrent, watchedNewPw, watchedConfirmPw] = useWatch({
    control,
    name: ["current", "newPw", "confirmPw"],
  });
  const canSubmit = !!watchedCurrent && !!watchedNewPw && !!watchedConfirmPw && isValid && !isSubmitting;

  useLayoutEffect(() => {
    reset({ current: "", newPw: "", confirmPw: "" });
  }, [open, reset]);
  useResetOnOpen(open, () => setSuccess(false));

  const onValid = async (values: ChangePasswordFormValues) => {
    try {
      await updatePw(values.current, values.newPw);
      setSuccess(true);
      setTimeout(onClose, 1200);
    } catch (err) {
      if (isSdkError(err) && err.status === 403) {
        setError("root", { message: "Current password is incorrect." });
      } else {
        setError("root", { message: "Failed to change password." });
      }
    }
  };

  const rootError = formState.errors.root?.message;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Change Password"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit(onValid)()}
            disabled={!canSubmit}
            loading={isSubmitting}
            icon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Change Password
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(onValid)(e)} className="space-y-5">
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
        {rootError && <p className="text-2xs text-accent-red">{rootError}</p>}
        {success && (
          <p className="text-2xs text-accent-green">
            Password changed successfully.
          </p>
        )}
      </form>
    </Drawer>
  );
}

/* ─── Page ─── */

export default function Profile() {
  const { name, username, email, recoveryEmail, mfaEnabled, fetchUser } =
    useAuthStore();

  const [editDrawerOpen, setEditDrawerOpen] = useState(false);
  const [pwDrawerOpen, setPwDrawerOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const config = getConfig();
  const isCloud = config.cloud;
  const isCommunity = !config.cloud && !config.enterprise;
  const [mfaEnableOpen, setMfaEnableOpen] = useState(false);
  const [mfaDisableOpen, setMfaDisableOpen] = useState(false);

  useEffect(() => {
    void fetchUser();
  }, [fetchUser]);

  if (!name && !username) {
    return <PageLoader label="Loading profile" padding="lg" />;
  }

  const openEdit = () => setEditDrawerOpen(true);

  return (
    <div>
      <PageHeader
        icon={<UserIcon className="w-6 h-6" />}
        overline="Account"
        title="Profile"
        description="Manage your account details and security settings"
      >
        <Button
          variant="secondary"
          onClick={openEdit}
          icon={<PencilSquareIcon className="w-4 h-4" />}
        >
          Edit Profile
        </Button>
      </PageHeader>

      <div className="space-y-6 animate-fade-in">
        {/* ── Profile ── */}
        <SettingsCard title="Profile">
          <SettingsRow
            icon={<UserIcon className="w-4 h-4" />}
            title="Name"
            description="Your display name"
          >
            <span className="text-sm font-mono text-text-secondary">
              {name}
            </span>
          </SettingsRow>

          <SettingsRow
            icon={<UserCircleIcon className="w-4 h-4" />}
            title="Username"
            description="Legacy login identifier. Use email to sign in instead."
            badge={
              <span className="px-1.5 py-0.5 text-3xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                Deprecated
              </span>
            }
          >
            <span className="text-sm font-mono text-text-secondary">
              {username}
            </span>
          </SettingsRow>

          <SettingsRow
            icon={<EnvelopeIcon className="w-4 h-4" />}
            title="Email"
            description="Used for login and account communications"
          >
            <span className="text-sm font-mono text-text-secondary">
              {email}
            </span>
          </SettingsRow>

          <SettingsRow
            icon={<AtSymbolIcon className="w-4 h-4" />}
            title="Recovery Email"
            description="Used for account recovery if you lose access"
          >
            <span className="text-sm font-mono text-text-secondary">
              {recoveryEmail || (
                <span className="text-text-muted italic font-sans">
                  Not set
                </span>
              )}
            </span>
          </SettingsRow>
        </SettingsCard>

        {/* ── Security ── */}
        <SettingsCard title="Security">
          <SettingsRow
            icon={<LockClosedIcon className="w-4 h-4" />}
            title="Password"
            description="Credentials used to authenticate into your account"
          >
            <Button
              variant="secondary"
              onClick={() => setPwDrawerOpen(true)}
            >
              Change Password
            </Button>
          </SettingsRow>

          {hasMfaSupport() ? (
            <SettingsRow
              icon={<ShieldCheckIcon className="w-4 h-4" />}
              title="Multi-Factor Authentication"
              description="Add an extra layer of security with TOTP-based 2FA. Recovery codes are shown once during setup."
              badge={
                mfaEnabled ? (
                  <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-green/10 text-accent-green border border-accent-green/20">
                    Enabled
                  </span>
                ) : null
              }
            >
              {mfaEnabled ? (
                <Button
                  variant="secondary"
                  onClick={() => setMfaDisableOpen(true)}
                >
                  Disable
                </Button>
              ) : (
                <Button
                  variant="secondary"
                  onClick={() => setMfaEnableOpen(true)}
                >
                  Enable MFA
                </Button>
              )}
            </SettingsRow>
          ) : (
            <SettingsRow
              icon={<ShieldCheckIcon className="w-4 h-4" />}
              title="Multi-Factor Authentication"
              description="Enhance your account security with TOTP-based 2FA"
              badge={
                <span className="px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                  Pro
                </span>
              }
            >
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
            </SettingsRow>
          )}
        </SettingsCard>

        {/* ── Danger Zone ── */}
        <SettingsCard title="Danger Zone" danger>
          <SettingsRow
            icon={<TrashIcon className="w-4 h-4 text-accent-red" />}
            title="Delete Account"
            description={
              isCloud
                ? "Permanently remove your account and all associated data."
                : "Account deletion requires CLI or Admin Console access."
            }
          >
            <Button
              size="sm"
              variant="dangerSoft"
              onClick={() => setDeleteDialogOpen(true)}
              data-test="delete-account-btn"
              className="hover:border-accent-red/40"
            >
              Delete
            </Button>
          </SettingsRow>
        </SettingsCard>
      </div>

      <EditProfileDrawer
        open={editDrawerOpen}
        onClose={() => setEditDrawerOpen(false)}
        currentName={name ?? ""}
        currentUsername={username ?? ""}
        currentEmail={email ?? ""}
        currentRecoveryEmail={recoveryEmail ?? ""}
      />
      <ChangePasswordDrawer
        open={pwDrawerOpen}
        onClose={() => setPwDrawerOpen(false)}
      />
      {isCloud ? (
        <DeleteAccountDialog
          key={String(deleteDialogOpen)}
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
        />
      ) : (
        <DeleteAccountWarningDialog
          open={deleteDialogOpen}
          onClose={() => setDeleteDialogOpen(false)}
          isCommunity={isCommunity}
        />
      )}
      <MfaEnableDrawer
        open={mfaEnableOpen}
        onClose={() => setMfaEnableOpen(false)}
        onSuccess={() => {
          setMfaEnableOpen(false);
          void fetchUser(); // Refresh to update mfaEnabled
        }}
        currentRecoveryEmail={recoveryEmail ?? null}
      />
      <MfaDisableDialog
        open={mfaDisableOpen}
        onClose={() => setMfaDisableOpen(false)}
        onSuccess={() => {
          setMfaDisableOpen(false);
          void fetchUser(); // Refresh to update mfaEnabled
        }}
      />
    </div>
  );
}
