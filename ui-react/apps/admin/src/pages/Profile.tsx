import { useEffect, useState, FormEvent } from "react";
import { useAuthStore } from "../stores/authStore";
import PageHeader from "../components/common/PageHeader";
import Drawer from "../components/common/Drawer";
import { AxiosError } from "axios";
import { LABEL, INPUT } from "../utils/styles";
import {
  UserIcon,
  PencilSquareIcon,
  CheckIcon,
  UserCircleIcon,
  EnvelopeIcon,
  LockClosedIcon,
  TrashIcon,
  AtSymbolIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import MfaEnableDrawer from "../components/mfa/MfaEnableDrawer";
import MfaDisableDialog from "../components/mfa/MfaDisableDialog";
import { hasMfaSupport } from "../utils/features";

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

function validateRecoveryEmail(v: string, primary: string): string | null {
  if (!v) return null;
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return "Invalid email format";
  if (v === primary) return "Must be different from your email";
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

/* ─── Edit Profile Drawer ─── */

function EditProfileDrawer({
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
    recoveryEmail !== currentRecoveryEmail
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
  const { name, username, email, recoveryEmail, mfaEnabled, fetchUser } = useAuthStore();

  const [editDrawerOpen, setEditDrawerOpen] = useState(false);
  const [pwDrawerOpen, setPwDrawerOpen] = useState(false);
  const [mfaEnableOpen, setMfaEnableOpen] = useState(false);
  const [mfaDisableOpen, setMfaDisableOpen] = useState(false);

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
                <button
                  onClick={() => setMfaDisableOpen(true)}
                  className="px-4 py-2 bg-hover-strong hover:bg-hover-strong text-text-primary border border-border hover:border-border-light rounded-lg text-sm font-medium transition-all"
                >
                  Disable
                </button>
              ) : (
                <button
                  onClick={() => setMfaEnableOpen(true)}
                  className="px-4 py-2 bg-hover-strong hover:bg-hover-strong text-text-primary border border-border hover:border-border-light rounded-lg text-sm font-medium transition-all"
                >
                  Enable MFA
                </button>
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
              <a
                href="https://www.shellhub.io/pricing"
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-2 bg-hover-strong hover:bg-hover-strong text-text-primary border border-border hover:border-border-light rounded-lg text-sm font-medium transition-all"
              >
                Upgrade
              </a>
            </SettingsRow>
          )}
        </SettingsCard>

        {/* ── Danger Zone ── */}
        <SettingsCard title="Danger Zone" danger>
          <SettingsRow
            icon={<TrashIcon className="w-4 h-4 text-accent-red" />}
            title="Delete Account"
            description="Permanently remove your account and all associated data. Only available on ShellHub Cloud."
          >
            <span className="inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20">
              Cloud Only
            </span>
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
      <MfaEnableDrawer
        open={mfaEnableOpen}
        onClose={() => setMfaEnableOpen(false)}
        onSuccess={() => {
          setMfaEnableOpen(false);
          fetchUser(); // Refresh to update mfaEnabled
        }}
        currentRecoveryEmail={recoveryEmail ?? null}
      />
      <MfaDisableDialog
        open={mfaDisableOpen}
        onClose={() => setMfaDisableOpen(false)}
        onSuccess={() => {
          setMfaDisableOpen(false);
          fetchUser(); // Refresh to update mfaEnabled
        }}
      />
    </div>
  );
}
