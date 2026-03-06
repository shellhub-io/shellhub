import { useEffect, useState, FormEvent } from "react";
import { useAuthStore } from "../stores/authStore";
import { useNamespacesStore } from "../stores/namespacesStore";
import PageHeader from "../components/common/PageHeader";
import Drawer from "../components/common/Drawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import CopyButton from "../components/common/CopyButton";
import { AxiosError } from "axios";
import { LABEL, INPUT } from "../utils/styles";
import { validateRecoveryEmail } from "./profile/validate";
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
  ShieldCheckIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";

const USERNAME_REGEX = /^[a-z0-9_.@-]+$/;

/* ─── Validation ─── */

function validateName(v: string): string | null {
  if (!v.trim()) return "Name is required";
  if (v.length > 64) return "Name must be at most 64 characters";
  return null;
}

function validateUsername(v: string): string | null {
  if (!v.trim()) return "Username is required";
  if (v.length > 32) return "Username must be at most 32 characters";
  if (v !== v.toLowerCase()) return "Username must be lowercase";
  if (v.includes(" ")) return "Username cannot contain spaces";
  if (!USERNAME_REGEX.test(v))
    return "Only lowercase letters, numbers, dots, underscores, @ and hyphens are allowed";
  return null;
}

function validateEmail(v: string): string | null {
  if (!v.trim()) return "Email is required";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return "Invalid email format";
  return null;
}

function validatePassword(v: string): string | null {
  if (v.length < 5) return "Password must be at least 5 characters";
  if (v.length > 32) return "Password must be at most 32 characters";
  return null;
}

/* ─── Settings Card ─── */

function SettingsCard({
  title,
  children,
  danger,
}: {
  title: string;
  children: React.ReactNode;
  danger?: boolean;
}) {
  return (
    <div
      className={`bg-card border rounded-xl overflow-hidden ${danger ? "border-accent-red/20 border-l-2 border-l-accent-red/40" : "border-border"}`}
    >
      <div
        className={`px-5 py-3.5 border-b ${danger ? "border-accent-red/10" : "border-border"}`}
      >
        <h3
          className={`text-sm font-semibold ${danger ? "text-accent-red" : "text-text-primary"}`}
        >
          {title}
        </h3>
      </div>
      <div className="divide-y divide-border">{children}</div>
    </div>
  );
}

/* ─── Settings Row ─── */

function SettingsRow({
  icon,
  title,
  description,
  badge,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  badge?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-6 px-5 py-4">
      <div className="flex items-start gap-3 min-w-0 flex-1">
        <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
          {icon}
        </span>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium text-text-primary">{title}</p>
            {badge}
          </div>
          <p className="text-2xs text-text-muted mt-0.5 leading-relaxed">
            {description}
          </p>
        </div>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
}

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
  const namespaces = useNamespacesStore((s) => s.namespaces);
  const [error, setError] = useState("");

  const isNamespaceOwner = namespaces.some((ns) => ns.owner === userId);


  const handleDelete = async () => {
    setError("");
    try {
      await deleteUser();
      window.location.replace("/v2/ui/login");
    } catch (err) {
      if (err instanceof AxiosError && err.response?.status === 403) {
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
  const namespaces = useNamespacesStore((s) => s.namespaces);

  const isNamespaceOwner = namespaces.some((ns) => ns.owner === userId);
  const deleteCommand = `./bin/cli user delete ${username ?? ""}`;

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-md mx-4 p-6 shadow-2xl animate-slide-up">
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
              In Enterprise instances, user accounts can only be deleted via
              the Admin Console. Please access your{" "}
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
          <button
            onClick={onClose}
            data-test="close-btn"
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
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
  const [name, setName] = useState(currentName);
  const [username, setUsername] = useState(currentUsername);
  const [email, setEmail] = useState(currentEmail);
  const [recoveryEmail, setRecoveryEmail] = useState(currentRecoveryEmail);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setName(currentName);
      setUsername(currentUsername);
      setEmail(currentEmail);
      setRecoveryEmail(currentRecoveryEmail);
      setError("");
    }
  }, [open, currentName, currentUsername, currentEmail, currentRecoveryEmail]);

  const nameError = name !== currentName ? validateName(name) : null;
  const usernameError =
    username !== currentUsername ? validateUsername(username) : null;
  const emailError = email !== currentEmail ? validateEmail(email) : null;
  const recoveryEmailError =
    recoveryEmail !== currentRecoveryEmail || email !== currentEmail
      ? validateRecoveryEmail(recoveryEmail, email)
      : null;

  const hasValidationErrors =
    !!nameError || !!usernameError || !!emailError || !!recoveryEmailError;

  const hasChanges =
    name !== currentName ||
    username !== currentUsername ||
    email !== currentEmail ||
    recoveryEmail !== currentRecoveryEmail;

  const canSubmit = hasChanges && !hasValidationErrors && !submitting;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setSubmitting(true);
    setError("");

    const data: {
      name?: string;
      username?: string;
      email?: string;
      recovery_email?: string;
    } = {};
    if (name !== currentName) data.name = name;
    if (username !== currentUsername) data.username = username;
    if (email !== currentEmail) data.email = email;
    if (recoveryEmail !== currentRecoveryEmail)
      data.recovery_email = recoveryEmail;

    try {
      await updateProfile(data);
      onClose();
    } catch (err) {
      if (err instanceof AxiosError) {
        if (err.response?.status === 409) {
          setError("That username or email is already in use.");
        } else if (err.response?.status === 400) {
          setError("Some fields have invalid values. Review and try again.");
        } else {
          setError("Failed to update profile.");
        }
      } else {
        setError("Failed to update profile.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Profile"
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
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <CheckIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Save
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className={LABEL}>Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus={open}
            className={INPUT}
            placeholder="Your name"
            maxLength={64}
          />
          <p className="text-2xs text-text-muted mt-1.5">1-64 characters</p>
          {nameError && (
            <p className="text-2xs text-accent-red mt-1">{nameError}</p>
          )}
        </div>
        <div>
          <div className="flex items-center gap-2 mb-1.5">
            <label className={`${LABEL} !mb-0`}>Username</label>
            <span className="px-1.5 py-0.5 text-3xs font-mono font-semibold uppercase tracking-wider rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
              Deprecated
            </span>
          </div>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value.toLowerCase())}
            className={INPUT}
            placeholder="username"
            maxLength={32}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            Lowercase letters, numbers, dots, underscores, @ and hyphens
          </p>
          {usernameError && (
            <p className="text-2xs text-accent-red mt-1">{usernameError}</p>
          )}
        </div>
        <div>
          <label className={LABEL}>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={INPUT}
            placeholder="you@example.com"
          />
          {emailError && (
            <p className="text-2xs text-accent-red mt-1">{emailError}</p>
          )}
        </div>
        <div>
          <label className={LABEL}>Recovery Email</label>
          <input
            type="email"
            value={recoveryEmail}
            onChange={(e) => setRecoveryEmail(e.target.value)}
            className={INPUT}
            placeholder="recovery@example.com"
          />
          <p className="text-2xs text-text-muted mt-1.5">
            Optional. Used for account recovery if you lose access.
          </p>
          {recoveryEmailError && (
            <p className="text-2xs text-accent-red mt-1">
              {recoveryEmailError}
            </p>
          )}
        </div>
        {error && <p className="text-2xs text-accent-red">{error}</p>}
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
  const [current, setCurrent] = useState("");
  const [newPw, setNewPw] = useState("");
  const [confirmPw, setConfirmPw] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (open) {
      setCurrent("");
      setNewPw("");
      setConfirmPw("");
      setError("");
      setSuccess(false);
    }
  }, [open]);

  const newPwError = newPw ? validatePassword(newPw) : null;
  const confirmError =
    confirmPw && newPw !== confirmPw ? "Passwords do not match" : null;
  const canSubmit =
    current &&
    newPw &&
    confirmPw &&
    !newPwError &&
    !confirmError &&
    !submitting;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setSubmitting(true);
    setError("");
    try {
      await updatePw(current, newPw);
      setSuccess(true);
      setTimeout(onClose, 1200);
    } catch (err) {
      if (err instanceof AxiosError && err.response?.status === 403) {
        setError("Current password is incorrect.");
      } else {
        setError("Failed to change password.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Change Password"
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
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <CheckIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Change Password
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className={LABEL}>Current Password</label>
          <input
            type="password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            autoFocus={open}
            className={INPUT}
            autoComplete="current-password"
          />
        </div>
        <div>
          <label className={LABEL}>New Password</label>
          <input
            type="password"
            value={newPw}
            onChange={(e) => setNewPw(e.target.value)}
            className={INPUT}
            autoComplete="new-password"
          />
          <p className="text-2xs text-text-muted mt-1.5">5-32 characters</p>
          {newPwError && (
            <p className="text-2xs text-accent-red mt-1">{newPwError}</p>
          )}
        </div>
        <div>
          <label className={LABEL}>Confirm New Password</label>
          <input
            type="password"
            value={confirmPw}
            onChange={(e) => setConfirmPw(e.target.value)}
            className={INPUT}
            autoComplete="new-password"
          />
          {confirmError && (
            <p className="text-2xs text-accent-red mt-1">{confirmError}</p>
          )}
        </div>
        {error && <p className="text-2xs text-accent-red">{error}</p>}
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
  const { name, username, email, recoveryEmail, fetchUser } = useAuthStore();

  const [editDrawerOpen, setEditDrawerOpen] = useState(false);
  const [pwDrawerOpen, setPwDrawerOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const config = getConfig();
  const isCloud = config.cloud;
  const isCommunity = !config.cloud && !config.enterprise;

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  if (!name && !username) {
    return (
      <div className="flex items-center justify-center py-32">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
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
        <button
          onClick={openEdit}
          className="px-4 py-2.5 bg-hover-strong hover:bg-hover-strong text-text-primary border border-border hover:border-border-light rounded-lg text-sm font-medium transition-all flex items-center gap-2"
        >
          <PencilSquareIcon className="w-4 h-4" />
          Edit Profile
        </button>
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
            <button
              onClick={() => setPwDrawerOpen(true)}
              className="px-4 py-2 bg-hover-strong hover:bg-hover-strong text-text-primary border border-border hover:border-border-light rounded-lg text-sm font-medium transition-all"
            >
              Change Password
            </button>
          </SettingsRow>
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
            <button
              onClick={() => setDeleteDialogOpen(true)}
              data-test="delete-account-btn"
              className="px-4 py-2 bg-accent-red/10 hover:bg-accent-red/20 text-accent-red border border-accent-red/20 hover:border-accent-red/40 rounded-lg text-sm font-medium transition-all"
            >
              Delete
            </button>
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
    </div>
  );
}
