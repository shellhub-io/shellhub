import { useState } from "react";
import {
  EnvelopeOpenIcon,
  PencilSquareIcon,
  ArrowPathIcon,
  TrashIcon,
  UserPlusIcon,
  ChevronDownIcon,
} from "@heroicons/react/24/outline";
import type { MembershipInvitation } from "@/client";
import { isSdkError } from "@/api/errors";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDateShort } from "@/utils/date";
import { useNamespaceInvitations } from "@/hooks/useInvitations";
import {
  useCancelMembershipInvitation,
  useGenerateInvitationLink,
} from "@/hooks/useInvitationMutations";
import {
  isInvitationExpired,
  type InvitationStatus,
} from "@/utils/invitations";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { ExpiredBadge, RoleBadge } from "./constants";
import InvitationDrawer from "./InvitationDrawer";
import EditInvitationDrawer from "./EditInvitationDrawer";
import { Button, IconButton } from "@shellhub/design-system/primitives";

const PER_PAGE = 10;

type StatusOption = { value: InvitationStatus; label: string };

const STATUS_OPTIONS: StatusOption[] = [
  { value: "pending", label: "Pending" },
  { value: "accepted", label: "Accepted" },
  { value: "rejected", label: "Rejected" },
  { value: "cancelled", label: "Cancelled" },
];

const STATUS_STYLES: Record<
  InvitationStatus,
  { bg: string; text: string; border: string; label: string }
> = {
  pending: {
    bg: "bg-accent-yellow/10",
    text: "text-accent-yellow",
    border: "border-accent-yellow/20",
    label: "Pending",
  },
  accepted: {
    bg: "bg-accent-green/10",
    text: "text-accent-green",
    border: "border-accent-green/20",
    label: "Accepted",
  },
  rejected: {
    bg: "bg-accent-red/10",
    text: "text-accent-red",
    border: "border-accent-red/20",
    label: "Rejected",
  },
  cancelled: {
    bg: "bg-hover-medium",
    text: "text-text-muted",
    border: "border-border",
    label: "Cancelled",
  },
};

function StatusBadge({ status }: { status: InvitationStatus }) {
  const style = STATUS_STYLES[status];
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${style.bg} ${style.text} ${style.border}`}
    >
      {style.label}
    </span>
  );
}

function rightColumnHeader(status: InvitationStatus): string {
  switch (status) {
    case "accepted":
      return "Accepted";
    case "rejected":
      return "Rejected";
    case "cancelled":
      return "Cancelled";
    case "pending":
      return "Expires";
  }
}

function resendEnabled(inv: MembershipInvitation): boolean {
  if (inv.status === "cancelled") return true;
  if (inv.status === "pending" && isInvitationExpired(inv.expires_at))
    return true;
  return false;
}

function cancelErrorMessage(err: unknown): string {
  if (isSdkError(err)) {
    switch (err.status) {
      case 403:
        return "You don't have permission to cancel invitations.";
      case 404:
        return "This invitation no longer exists.";
    }
  }
  return "Failed to cancel the invitation. Please try again.";
}

function resendErrorMessage(err: unknown): string {
  if (isSdkError(err)) {
    switch (err.status) {
      case 403:
        return "You don't have permission to resend invitations.";
      case 404:
        return "This invitation no longer exists.";
      case 409:
        return "This user is already a member or has a pending invitation.";
    }
  }
  return "Failed to resend the invitation. Please try again.";
}

const STATUS_ALLOWLIST: readonly InvitationStatus[] = [
  "pending",
  "accepted",
  "rejected",
  "cancelled",
];

type InvListParams = {
  page: number;
  status: InvitationStatus;
};

const INV_LIST_DEFAULTS: InvListParams = { page: 1, status: "pending" };
const INV_LIST_CONSTRAINTS: { status: readonly InvitationStatus[] } = {
  status: STATUS_ALLOWLIST,
};

function InvitationsTab({ tenantId }: { tenantId: string }) {
  const {
    params,
    setPage,
    setFilter,
  } = usePaginatedListState<InvListParams>({
    prefix: "inv",
    defaults: INV_LIST_DEFAULTS,
    constraints: INV_LIST_CONSTRAINTS,
  });

  const status = params.status;
  const page = params.page;

  const [inviteOpen, setInviteOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [cancelTarget, setCancelTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [resendTarget, setResendTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [resendError, setResendError] = useState<string | null>(null);

  const { invitations, totalCount, isLoading } = useNamespaceInvitations({
    tenantId,
    status,
    page,
    perPage: PER_PAGE,
  });

  const cancelInvitation = useCancelMembershipInvitation();
  const resendInvitation = useGenerateInvitationLink();

  const totalPages = Math.max(1, Math.ceil(totalCount / PER_PAGE));

  const handleStatusChange = (next: InvitationStatus) => {
    setFilter("status", next);
  };

  const rightHeader = rightColumnHeader(status);

  const columns: Column<MembershipInvitation>[] = [
    {
      key: "email",
      header: "Email",
      render: (inv) => (
        <span className="text-sm font-medium text-text-primary">
          {inv.user.email}
        </span>
      ),
    },
    {
      key: "role",
      header: "Role",
      render: (inv) => <RoleBadge role={inv.role} />,
    },
    {
      key: "status",
      header: "Status",
      render: (inv) => {
        const expired =
          inv.status === "pending" && isInvitationExpired(inv.expires_at);
        return (
          <div className="flex items-center gap-1.5">
            <StatusBadge status={inv.status} />
            {expired && <ExpiredBadge />}
          </div>
        );
      },
    },
    {
      key: "created",
      header: "Created",
      render: (inv) => (
        <span className="text-xs text-text-secondary">
          {formatDateShort(inv.created_at)}
        </span>
      ),
    },
    {
      key: "timestamp",
      header: rightHeader,
      render: (inv) => {
        if (inv.status === "pending") {
          if (!inv.expires_at) {
            return <span className="text-xs text-text-muted">Never</span>;
          }
          const expired = isInvitationExpired(inv.expires_at);
          return (
            <span
              className={`text-xs ${expired ? "text-accent-red" : "text-text-secondary"}`}
            >
              {formatDateShort(inv.expires_at)}
            </span>
          );
        }
        return (
          <span className="text-xs text-text-secondary">
            {formatDateShort(inv.status_updated_at)}
          </span>
        );
      },
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (inv) => {
        const canEdit = inv.status === "pending";
        const canCancel = inv.status === "pending";
        const canResend = resendEnabled(inv);
        return (
          <div className="flex items-center justify-end gap-1">
            <RestrictedAction action="namespace:editInvitation">
              <IconButton
                variant="primary"
                disabled={!canEdit}
                title={
                  canEdit
                    ? "Edit role"
                    : "Only pending invitations can be edited"
                }
                aria-label="Edit invitation role"
                onClick={() => canEdit && setEditTarget(inv)}
              >
                <PencilSquareIcon className="w-4 h-4" />
              </IconButton>
            </RestrictedAction>
            <RestrictedAction action="namespace:addMember">
              <IconButton
                variant="ghost"
                disabled={!canResend}
                title={
                  canResend
                    ? "Resend invitation"
                    : "Only cancelled or expired invitations can be resent"
                }
                aria-label="Resend invitation"
                onClick={() => canResend && setResendTarget(inv)}
              >
                <ArrowPathIcon className="w-4 h-4" />
              </IconButton>
            </RestrictedAction>
            <RestrictedAction action="namespace:cancelInvitation">
              <IconButton
                variant="danger"
                disabled={!canCancel}
                title={
                  canCancel
                    ? "Cancel invitation"
                    : "Only pending invitations can be cancelled"
                }
                aria-label="Cancel invitation"
                onClick={() => canCancel && setCancelTarget(inv)}
              >
                <TrashIcon className="w-4 h-4" />
              </IconButton>
            </RestrictedAction>
          </div>
        );
      },
    },
  ];

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5 gap-3 flex-wrap">
        <div className="flex items-center gap-3">
          <p className="text-sm text-text-muted">
            {totalCount} invitation
            {totalCount !== 1 ? "s" : ""}
          </p>
          <div className="relative">
            <select
              value={status}
              onChange={(e) =>
                handleStatusChange(e.target.value as InvitationStatus)
              }
              className="appearance-none pl-3 pr-8 h-8 bg-card border border-border rounded-md text-xs font-medium text-text-secondary hover:border-border-light focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all cursor-pointer"
              aria-label="Filter invitations by status"
            >
              {STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <ChevronDownIcon
              className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-3 h-3 text-text-muted"
              strokeWidth={2.5}
            />
          </div>
        </div>
        <RestrictedAction action="namespace:addMember">
          <Button
            onClick={() => setInviteOpen(true)}
            icon={<UserPlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Invite Member
          </Button>
        </RestrictedAction>
      </div>

      <DataTable
        columns={columns}
        data={invitations}
        rowKey={(inv) => `${inv.status}-${inv.user.id}`}
        isLoading={isLoading}
        loadingMessage="Loading invitations..."
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
        rowClassName={(inv) =>
          inv.status === "pending" && isInvitationExpired(inv.expires_at)
            ? "bg-accent-red/[0.03] border-l-2 border-l-accent-red/50"
            : "border-l-2 border-l-transparent"
        }
        emptyState={
          <div className="text-center">
            <EnvelopeOpenIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-sm text-text-muted">
              No {STATUS_STYLES[status].label.toLowerCase()} invitations
            </p>
            <p className="text-2xs text-text-muted/60 mt-1">
              {status === "pending"
                ? "Invite teammates to collaborate in this namespace"
                : "Switch filters to see other invitations"}
            </p>
          </div>
        }
      />

      <InvitationDrawer
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        tenantId={tenantId}
      />

      <EditInvitationDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        tenantId={tenantId}
        invitation={editTarget}
      />

      <ConfirmDialog
        open={!!cancelTarget}
        onClose={() => {
          setCancelTarget(null);
          setCancelError(null);
        }}
        onConfirm={async () => {
          if (!cancelTarget) return;
          setCancelError(null);
          // Drop one page *before* awaiting so the refetch targeted by
          // useInvalidateByIds hits the correct page after the row disappears.
          const shouldStepBack = invitations.length === 1 && page > 1;
          if (shouldStepBack) setPage(page - 1);
          try {
            await cancelInvitation.mutateAsync({
              path: {
                tenant: tenantId,
                "user-id": cancelTarget.user.id,
              },
            });
            setCancelTarget(null);
          } catch (err) {
            if (shouldStepBack) setPage(page);
            setCancelError(cancelErrorMessage(err));
          }
        }}
        title="Cancel Invitation"
        description={
          <>
            Cancel the invitation sent to{" "}
            <span className="font-medium text-text-primary">
              {cancelTarget?.user.email}
            </span>
            ? They will no longer be able to join via the existing link.
          </>
        }
        confirmLabel="Cancel Invitation"
        cancelLabel="Keep"
        variant="danger"
        errorMessage={cancelError}
      />

      <ConfirmDialog
        open={!!resendTarget}
        onClose={() => {
          setResendTarget(null);
          setResendError(null);
        }}
        onConfirm={async () => {
          if (!resendTarget) return;
          setResendError(null);
          try {
            await resendInvitation.mutateAsync({
              path: { tenant: tenantId },
              body: {
                email: resendTarget.user.email,
                role: resendTarget.role,
              },
            });
            setResendTarget(null);
          } catch (err) {
            setResendError(resendErrorMessage(err));
          }
        }}
        title="Resend Invitation"
        description={
          <>
            Send a fresh invitation to{" "}
            <span className="font-medium text-text-primary">
              {resendTarget?.user.email}
            </span>
            ? A new email will be dispatched and any previous link will become
            invalid.
          </>
        }
        confirmLabel="Resend"
        variant="primary"
        errorMessage={resendError}
      />
    </div>
  );
}

export default InvitationsTab;
