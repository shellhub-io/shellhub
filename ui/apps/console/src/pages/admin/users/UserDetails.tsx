import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  UsersIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
  InformationCircleIcon,
  ClockIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useAdminUser } from "@/hooks/useAdminUsers";
import Breadcrumb from "@/components/common/Breadcrumb";
import { useLoginAsUser } from "@/hooks/useLoginAsUser";
import UserStatusChip from "./UserStatusChip";
import EditUserDrawer from "./EditUserDrawer";
import ResetPasswordDialog from "./ResetPasswordDialog";
import DeleteUserDialog from "./DeleteUserDialog";
import { formatDateFull } from "@/utils/date";
import InfoItem from "@/components/common/InfoItem";
import PageLoader from "@/components/common/PageLoader";
import {
  Badge,
  Button,
  Card,
  IconButton,
} from "@shellhub/design-system/primitives";

const ZERO_DATE = "0001-01-01T00:00:00Z";

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
    return <PageLoader label="Loading user details" />;
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

  const isSamlOnly =
    user.preferences.auth_methods.length === 1 &&
    user.preferences.auth_methods[0] === "saml";
  const userStatus = user.status;
  const lastLogin = user.last_login === ZERO_DATE ? null : user.last_login;

  return (
    <div className="animate-fade-in">
      <Breadcrumb
        items={[
          { label: "Users", to: "/admin/users" },
          { label: user.username },
        ]}
      />

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
              {user.admin && <Badge color="yellow">Admin</Badge>}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="secondary"
            icon={<PencilSquareIcon className="w-4 h-4" />}
            onClick={() => setEditOpen(true)}
          >
            Edit
          </Button>
          {isSamlOnly && (
            <Button
              variant="secondary"
              icon={<KeyIcon className="w-4 h-4" />}
              title="Enable local authentication for this SAML-only user"
              onClick={() => setResetPasswordOpen(true)}
            >
              Set Password
            </Button>
          )}
          <Button
            variant={loginAsErrorId === id ? "destructive" : "primary"}
            icon={<ArrowRightStartOnRectangleIcon className="w-4 h-4" />}
            loading={loginAsId === id}
            disabled={loginAsId === id}
            onClick={() => id && void loginAs(id)}
          >
            {loginAsErrorId === id ? "Retry Login" : "Login as User"}
          </Button>
          <IconButton
            variant="danger"
            size="lg"
            title="Delete user"
            aria-label={`Delete ${user.name}`}
            className="border border-border"
            onClick={() => setDeleteOpen(true)}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {/* Identity Card */}
        <Card className="p-5 space-y-4">
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
            <InfoItem label="Status">
              <UserStatusChip status={userStatus} />
            </InfoItem>
          </dl>
        </Card>

        {/* Account Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Account
          </h3>
          <dl className="space-y-3">
            <InfoItem label="Created" value={formatDateFull(user.created_at)} />
            <InfoItem
              label="Last Login"
              value={lastLogin ? formatDateFull(lastLogin) : "Never logged in"}
            />
            <InfoItem
              label="Max Namespaces"
              value={formatMaxNamespaces(user.max_namespaces)}
            />
            <InfoItem
              label="Namespaces Owned"
              value={String(user.namespacesOwned)}
            />
            <InfoItem label="MFA">
              <span
                className={cn(
                  "inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md",
                  user.mfa.enabled
                    ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                    : "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20",
                )}
              >
                {user.mfa.enabled ? "Enabled" : "Disabled"}
              </span>
            </InfoItem>
            <InfoItem label="Auth Methods">
              {user.preferences.auth_methods.map((method) => (
                <Badge key={method} color="primary">
                  {method.toUpperCase()}
                </Badge>
              ))}
            </InfoItem>
          </dl>
        </Card>
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
