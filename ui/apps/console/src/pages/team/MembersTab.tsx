import { lazy, Suspense, useState } from "react";
import {
  PlusIcon,
  UserGroupIcon,
  PencilSquareIcon,
  TrashIcon,
  CheckIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import type { MemberView, MembershipInvitation } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import {
  useNamespaceMembers,
  type NamespaceMember,
} from "@/hooks/useNamespaces";
import { useNamespaceInvitations } from "@/hooks/useInvitations";
import { useRemoveMember, useApproveMember } from "@/hooks/useMemberMutations";
import {
  useCancelMembershipInvitation,
  useGenerateInvitationLink,
} from "@/hooks/useInvitationMutations";
import { isSdkError } from "@/api/errors";
import { isInvitationExpired } from "@/utils/invitations";
import { formatDateShort } from "@/utils/date";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import CopyButton from "@/components/common/CopyButton";
import DataTable, { type Column } from "@/components/common/DataTable";
import { RoleBadge } from "./constants";
import UserBadge from "@/components/common/UserBadge";
import EditMemberDrawer from "./EditMemberDrawer";
import RestrictedAction from "@/components/common/RestrictedAction";

// Cloud/enterprise add-member drawer — lazy so its transitive deps (CopyButton,
// isSdkError, etc.) don't ship to the community bundle.
const AddMemberDrawer = lazy(() => import("./AddMemberDrawer"));

// A single table row is either a real member or a pending invitation (the
// invitee hasn't completed their account yet). Merging the two sources here —
// not in the backend — keeps each endpoint honest per edition; the unified
// table is purely presentation.
type Row =
  | { kind: "member"; key: string; member: MemberView }
  | { kind: "invite"; key: string; invite: MembershipInvitation };

function Badge({
  tone,
  children,
}: {
  tone: "blue" | "yellow" | "green";
  children: React.ReactNode;
}) {
  const styles = {
    blue: "bg-accent-blue/10 text-accent-blue border-accent-blue/20",
    yellow: "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
    green: "bg-accent-green/10 text-accent-green border-accent-green/20",
  }[tone];
  return (
    <span
      className={cn(
        "inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border",
        styles,
      )}
    >
      {children}
    </span>
  );
}

function memberToNamespaceMember(m: MemberView): NamespaceMember {
  return {
    id: m.id ?? "",
    role: m.role ?? "observer",
    email: m.email ?? "",
    added_at: m.added_at,
  };
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

function MembersTab({ tenantId }: { tenantId: string }) {
  const { members: memberViews, isLoading: membersLoading } =
    useNamespaceMembers(tenantId);

  // Pending invitations exist in every edition, so always fetch them.
  const { invitations } = useNamespaceInvitations({
    tenantId,
    status: "pending",
    perPage: 100,
    enabled: true,
  });

  const removeMember = useRemoveMember();
  const approveMember = useApproveMember();
  const cancelInvitation = useCancelMembershipInvitation();
  const regenerateInvitation = useGenerateInvitationLink();

  const currentUserEmail = useAuthStore((s) => s.email);
  // Approving a provisioned account is an instance-admin (superadmin) act. A
  // namespace owner/admin who isn't a superadmin only sees the status.
  const isSuperAdmin = useAuthStore((s) => s.isAdmin);

  const [addOpen, setAddOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<NamespaceMember | null>(null);
  const [removeTarget, setRemoveTarget] = useState<MemberView | null>(null);
  const [cancelTarget, setCancelTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [regenTarget, setRegenTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [removeError, setRemoveError] = useState<string | null>(null);
  const [cancelError, setCancelError] = useState<string | null>(null);
  const [regenError, setRegenError] = useState<string | null>(null);

  const closeRemove = () => {
    setRemoveError(null);
    setRemoveTarget(null);
  };

  const confirmRemove = async () => {
    if (!removeTarget?.id) return;
    setRemoveError(null);
    try {
      await removeMember.mutateAsync({
        path: { tenant: tenantId, uid: removeTarget.id },
      });
      closeRemove();
    } catch (err) {
      setRemoveError(
        err instanceof Error ? err.message : "Failed to remove member.",
      );
    }
  };

  // Pending invitations first (they need action), then real members sorted by
  // email. The owner is shown too (complete roster) but has no destructive
  // actions — ownership is transferred, not removed here.
  const members = memberViews
    .filter((m) => !!m.id && !!m.email)
    .sort((a, b) => (a.email ?? "").localeCompare(b.email ?? ""));

  const rows: Row[] = [
    ...invitations.map((invite): Row => ({
      kind: "invite",
      key: `invite-${invite.user.id}`,
      invite,
    })),
    ...members.map((member): Row => ({
      kind: "member",
      key: `member-${member.id}`,
      member,
    })),
  ];

  const columns: Column<Row>[] = [
    {
      key: "member",
      header: "Member",
      render: (row) => {
        const isInvite = row.kind === "invite";
        const name = isInvite ? "" : (row.member.name ?? "");
        const email = isInvite
          ? row.invite.user.email
          : (row.member.email ?? "");
        const isSelf = !isInvite && row.member.email === currentUserEmail;
        return (
          <UserBadge
            name={name}
            email={email}
            trailing={
              isSelf ? (
                <span className="text-2xs text-text-muted font-mono">
                  (you)
                </span>
              ) : undefined
            }
          />
        );
      },
    },
    {
      key: "role",
      header: "Role",
      render: (row) => (
        <RoleBadge
          role={
            row.kind === "invite"
              ? row.invite.role
              : (row.member.role ?? "observer")
          }
        />
      ),
    },
    {
      key: "status",
      header: "Status",
      render: (row) => {
        if (row.kind === "invite") {
          const expired = isInvitationExpired(row.invite.expires_at);
          return (
            <div className="flex flex-col items-start gap-1">
              <Badge tone="yellow">Pending invite</Badge>
              {row.invite.expires_at && (
                <span
                  className={cn(
                    "text-2xs",
                    expired ? "text-accent-red" : "text-text-muted",
                  )}
                >
                  {expired
                    ? "expired"
                    : `expires ${formatDateShort(row.invite.expires_at)}`}
                </span>
              )}
            </div>
          );
        }
        switch (row.member.status) {
          case "awaiting_approval":
            return <Badge tone="blue">Awaiting approval</Badge>;
          case "not-confirmed":
            return <Badge tone="yellow">Pending activation</Badge>;
          default:
            return <Badge tone="green">Active</Badge>;
        }
      },
    },
    {
      key: "joined",
      header: "Added",
      render: (row) => {
        // Members show when they joined; pending invitees show when they were
        // invited (they haven't joined yet).
        const date =
          row.kind === "invite" ? row.invite.created_at : row.member.added_at;
        return (
          <span className="text-xs text-text-secondary">
            {date ? formatDateShort(date) : "—"}
          </span>
        );
      },
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (row) => {
        if (row.kind === "invite") {
          const inv = row.invite;
          // The backend only regenerates an expired invitation; a still-valid one 409s (copy the
          // existing link instead). Gate the action to match, as the old InvitationsTab did.
          const expired = isInvitationExpired(inv.expires_at);
          return (
            <div className="flex items-center justify-end gap-1">
              {inv.invite_url && (
                <RestrictedAction action="namespace:addMember">
                  <CopyButton text={inv.invite_url} />
                </RestrictedAction>
              )}
              <RestrictedAction action="namespace:addMember">
                <IconButton
                  variant="ghost"
                  disabled={!expired}
                  title={
                    expired
                      ? "Regenerate link (invalidates the current one)"
                      : "The current link is still valid; copy it instead"
                  }
                  aria-label="Regenerate invitation link"
                  onClick={() => setRegenTarget(inv)}
                >
                  <ArrowPathIcon className="w-4 h-4" />
                </IconButton>
              </RestrictedAction>
              <RestrictedAction action="namespace:cancelInvitation">
                <IconButton
                  variant="danger"
                  title="Cancel invitation"
                  aria-label="Cancel invitation"
                  onClick={() => setCancelTarget(inv)}
                >
                  <TrashIcon className="w-4 h-4" />
                </IconButton>
              </RestrictedAction>
            </div>
          );
        }

        const m = row.member;
        if (m.email === currentUserEmail || m.role === "owner") return null;
        return (
          <div className="flex items-center justify-end gap-1">
            {m.status === "awaiting_approval" && isSuperAdmin && (
              <IconButton
                variant="primary"
                title="Approve account"
                aria-label="Approve account"
                loading={approveMember.isPending}
                onClick={() =>
                  m.id && approveMember.mutate({ path: { id: m.id } })
                }
              >
                <CheckIcon className="w-4 h-4" />
              </IconButton>
            )}
            <RestrictedAction action="namespace:editMember">
              <IconButton
                variant="primary"
                title="Edit role"
                aria-label="Edit role"
                onClick={() => setEditTarget(memberToNamespaceMember(m))}
              >
                <PencilSquareIcon className="w-4 h-4" />
              </IconButton>
            </RestrictedAction>
            <RestrictedAction action="namespace:removeMember">
              <IconButton
                variant="danger"
                title="Remove"
                aria-label="Remove member"
                onClick={() => setRemoveTarget(m)}
              >
                <TrashIcon className="w-4 h-4" />
              </IconButton>
            </RestrictedAction>
          </div>
        );
      },
    },
  ];

  const summary = [
    `${members.length} member${members.length !== 1 ? "s" : ""}`,
    invitations.length > 0 ? `${invitations.length} pending` : null,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">{summary}</p>
        <RestrictedAction action="namespace:addMember">
          <Button
            onClick={() => setAddOpen(true)}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Member
          </Button>
        </RestrictedAction>
      </div>

      <DataTable
        columns={columns}
        data={rows}
        rowKey={(r) => r.key}
        isLoading={membersLoading}
        loadingMessage="Loading members..."
        rowClassName={(r) =>
          r.kind === "invite"
            ? "border-l-2 border-l-accent-yellow/40"
            : "border-l-2 border-l-transparent"
        }
        emptyState={
          <div className="text-center">
            <UserGroupIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-sm text-text-muted">No members yet</p>
            <p className="text-2xs text-text-muted/60 mt-1">
              Add members to collaborate in this namespace
            </p>
          </div>
        }
      />

      <Suspense fallback={null}>
        <AddMemberDrawer
          open={addOpen}
          onClose={() => setAddOpen(false)}
          tenantId={tenantId}
        />
      </Suspense>

      <EditMemberDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        tenantId={tenantId}
        member={editTarget}
      />

      <ConfirmDialog
        open={!!removeTarget}
        onClose={closeRemove}
        onConfirm={confirmRemove}
        title="Remove Member"
        description={
          <>
            Are you sure you want to remove{" "}
            <span className="font-medium text-text-primary">
              {removeTarget?.email}
            </span>{" "}
            from this namespace?
          </>
        }
        confirmLabel="Remove"
      >
        {removeError && (
          <p className="text-xs text-accent-red">{removeError}</p>
        )}
      </ConfirmDialog>

      <ConfirmDialog
        open={!!cancelTarget}
        onClose={() => {
          setCancelTarget(null);
          setCancelError(null);
        }}
        onConfirm={async () => {
          if (!cancelTarget) return;
          setCancelError(null);
          try {
            await cancelInvitation.mutateAsync({
              path: { tenant: tenantId, "user-id": cancelTarget.user.id },
            });
            setCancelTarget(null);
          } catch (err) {
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
        open={!!regenTarget}
        onClose={() => {
          setRegenTarget(null);
          setRegenError(null);
        }}
        onConfirm={async () => {
          if (!regenTarget) return;
          setRegenError(null);
          try {
            await regenerateInvitation.mutateAsync({
              path: { tenant: tenantId },
              body: {
                email: regenTarget.user.email,
                role: regenTarget.role,
              },
            });
            setRegenTarget(null);
          } catch {
            setRegenError("Failed to regenerate the link. Please try again.");
          }
        }}
        title="Regenerate Link"
        description={
          <>
            Generate a fresh invitation link for{" "}
            <span className="font-medium text-text-primary">
              {regenTarget?.user.email}
            </span>
            ? The current link will stop working.
          </>
        }
        confirmLabel="Regenerate"
        variant="primary"
        errorMessage={regenError}
      />
    </div>
  );
}

export default MembersTab;
