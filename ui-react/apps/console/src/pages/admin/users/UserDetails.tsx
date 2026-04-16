import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ChevronRightIcon,
  UsersIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
  InformationCircleIcon,
  ClockIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import { useAdminUser } from "@/hooks/useAdminUsers";
import { useLoginAsUser } from "@/hooks/useLoginAsUser";
import CopyButton from "@/components/common/CopyButton";
import UserStatusChip from "./UserStatusChip";
import EditUserDrawer from "./EditUserDrawer";
import ResetPasswordDialog from "./ResetPasswordDialog";
import DeleteUserDialog from "./DeleteUserDialog";
import { formatDateFull } from "@/utils/date";

const LABEL
  = "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";
const ZERO_DATE = "0001-01-01T00:00:00Z";

function InfoItem({
  label,
  value,
  mono,
  copyable,
}: {
  label: string;
  value: string;
  mono?: boolean;
  copyable?: boolean;
}) {
  return (
    <div>
      <dt className={LABEL}>{label}</dt>
      <dd className="flex items-center gap-1 mt-0.5">
        <span
          className={`text-sm text-text-primary ${mono ? "font-mono text-xs" : "font-medium"}`}
        >
          {value || "\u2014"}
        </span>
        {copyable && value && <CopyButton text={value} />}
      </dd>
    </div>
  );
}

function formatMaxNamespaces(value: number): string {
  if (value < 0) return "Unlimited";
  if (value === 0) return "Disabled";
  return String(value);
}

export default function UserDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading, error } = useAdminUser(id ?? "");
  const user = data;

  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [resetPasswordOpen, setResetPasswordOpen] = useState(false);
  const {
    loginAs,
    loadingId: loginAsId,
    errorId: loginAsErrorId,
  } = useLoginAsUser();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24" role="status">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading user details</span>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="text-center py-24">
        <UsersIcon
          className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
          strokeWidth={1}
        />
        <p className="text-sm text-text-muted mb-2">User not found</p>
        <Link
          to="/admin/users"
          className="text-sm text-primary hover:underline"
        >
          Back to users
        </Link>
      </div>
    );
  }

  const isSamlOnly
    = user.preferences.auth_methods.length === 1
      && user.preferences.auth_methods[0] === "saml";
  const userStatus = user.status;
  const lastLogin = user.last_login === ZERO_DATE ? null : user.last_login;

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/users"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Users
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary">
          {user.username}
        </span>
      </nav>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
            <UsersIcon className="w-7 h-7 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              {user.name}
            </h1>
            <div className="flex items-center gap-2 mt-1.5">
              <UserStatusChip status={userStatus} />
              {user.admin && (
                <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                  Admin
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => setEditOpen(true)}
            className="flex items-center gap-2 px-4 py-2.5 border border-border text-text-secondary hover:text-text-primary hover:border-border-light rounded-lg text-sm font-semibold transition-all"
          >
            <PencilSquareIcon className="w-4 h-4" />
            Edit
          </button>
          {isSamlOnly && (
            <button
              onClick={() => setResetPasswordOpen(true)}
              className="flex items-center gap-2 px-4 py-2.5 border border-border text-text-secondary hover:text-text-primary hover:border-border-light rounded-lg text-sm font-semibold transition-all"
              title="Enable local authentication for this SAML-only user"
            >
              <KeyIcon className="w-4 h-4" />
              Set Password
            </button>
          )}
          <button
            onClick={() => id && void loginAs(id)}
            disabled={loginAsId === id}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all ${
              loginAsErrorId === id
                ? "bg-accent-red hover:bg-accent-red/80 text-white"
                : "bg-primary hover:bg-primary-600 text-white"
            }`}
          >
            {loginAsId === id ? (
              <span
                aria-hidden="true"
                className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              />
            ) : (
              <ArrowRightStartOnRectangleIcon className="w-4 h-4" />
            )}
            {loginAsErrorId === id ? "Retry Login" : "Login as User"}
          </button>
          <button
            onClick={() => setDeleteOpen(true)}
            className="p-2.5 rounded-lg text-text-muted hover:text-accent-red hover:bg-accent-red/10 border border-border transition-all"
            title="Delete user"
            aria-label={`Delete ${user.name}`}
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {/* Identity Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Identity
          </h3>
          <dl className="space-y-3">
            <InfoItem label="ID" value={user.id} mono copyable />
            <InfoItem label="Name" value={user.name} />
            <InfoItem label="Username" value={user.username} mono />
            <InfoItem label="Email" value={user.email} />
            <InfoItem
              label="Recovery Email"
              value={user.recovery_email ?? ""}
            />
            <div>
              <dt className={LABEL}>Status</dt>
              <dd className="mt-1">
                <UserStatusChip status={userStatus} />
              </dd>
            </div>
          </dl>
        </div>

        {/* Account Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Account
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Created</dt>
              <dd className={VALUE}>{formatDateFull(user.created_at)}</dd>
            </div>
            <div>
              <dt className={LABEL}>Last Login</dt>
              <dd className={VALUE}>
                {lastLogin ? formatDateFull(lastLogin) : "Never logged in"}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Max Namespaces</dt>
              <dd className={VALUE}>
                {formatMaxNamespaces(user.max_namespaces)}
              </dd>
            </div>
            <InfoItem
              label="Namespaces Owned"
              value={String(user.namespacesOwned)}
            />
            <div>
              <dt className={LABEL}>MFA</dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md ${
                    user.mfa.enabled
                      ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                      : "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20"
                  }`}
                >
                  {user.mfa.enabled ? "Enabled" : "Disabled"}
                </span>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Auth Methods</dt>
              <dd className="flex items-center gap-1.5 mt-1">
                {user.preferences.auth_methods.map((method) => (
                  <span
                    key={method}
                    className="inline-flex items-center px-2 py-0.5 text-2xs font-medium rounded-md bg-primary/10 text-primary border border-primary/20"
                  >
                    {method.toUpperCase()}
                  </span>
                ))}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Edit Drawer */}
      <EditUserDrawer
        open={editOpen}
        onClose={() => setEditOpen(false)}
        user={user}
      />

      {/* Reset Password Dialog */}
      <ResetPasswordDialog
        open={resetPasswordOpen}
        onClose={() => setResetPasswordOpen(false)}
        userId={id ?? ""}
      />

      {/* Delete Confirmation */}
      <DeleteUserDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        user={user ? { id: user.id, name: user.name } : null}
        onDeleted={() => void navigate("/admin/users")}
      />
    </div>
  );
}
