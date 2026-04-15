import { useState, useRef } from "react";
import { Link } from "react-router-dom";
import {
  EnvelopeIcon,
  InboxIcon,
  CheckIcon,
  XMarkIcon,
  ArrowRightIcon,
} from "@heroicons/react/24/outline";
import { getConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { useUserInvitations } from "@/hooks/useInvitations";
import {
  useAcceptInvite,
  useDeclineInvite,
} from "@/hooks/useInvitationMutations";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import type { MembershipInvitation } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { formatRelative } from "@/utils/date";
import { RoleBadge } from "@/pages/team/constants";
import { isInvitationExpired } from "@/utils/invitations";

function InvitationCard({
  invitation,
  onAccept,
  onDecline,
}: {
  invitation: MembershipInvitation;
  onAccept: (inv: MembershipInvitation) => void;
  onDecline: (inv: MembershipInvitation) => void;
}) {
  const expired = isInvitationExpired(invitation.expires_at);
  return (
    <li className="px-3.5 py-3 border-b border-border/60 last:border-b-0">
      <div className="flex items-start justify-between gap-2 mb-1.5">
        <p className="text-sm font-semibold text-text-primary truncate">
          {invitation.namespace.name}
        </p>
        <RoleBadge role={invitation.role} />
      </div>
      <p className="text-2xs text-text-muted font-mono truncate">
        Invited by {invitation.invited_by}
      </p>
      <p className="text-2xs text-text-muted mt-0.5">
        {expired ? (
          <span className="text-accent-red font-medium">Expired</span>
        ) : invitation.expires_at ? (
          <>Expires {formatRelative(invitation.expires_at)}</>
        ) : (
          <>No expiration</>
        )}
      </p>
      <div className="flex items-center gap-1.5 mt-2.5">
        <button
          onClick={() => onAccept(invitation)}
          disabled={expired}
          className="flex-1 inline-flex items-center justify-center gap-1 px-2.5 py-1.5 bg-primary/10 hover:bg-primary/20 border border-primary/20 text-primary rounded-md text-2xs font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-colors"
        >
          <CheckIcon className="w-3 h-3" strokeWidth={2.5} />
          Accept
        </button>
        <button
          onClick={() => onDecline(invitation)}
          className="flex-1 inline-flex items-center justify-center gap-1 px-2.5 py-1.5 bg-hover-subtle hover:bg-hover-medium border border-border text-text-secondary hover:text-text-primary rounded-md text-2xs font-semibold transition-colors"
        >
          <XMarkIcon className="w-3 h-3" strokeWidth={2.5} />
          Decline
        </button>
      </div>
    </li>
  );
}

export default function InvitationsMenu() {
  const isCloud = getConfig().cloud;
  const token = useAuthStore((s) => s.token);

  const [open, setOpen] = useState(false);
  const [acceptTarget, setAcceptTarget] = useState<MembershipInvitation | null>(
    null,
  );
  const [declineTarget, setDeclineTarget] =
    useState<MembershipInvitation | null>(null);
  const [acceptError, setAcceptError] = useState<string | null>(null);
  const [declineError, setDeclineError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useClickOutside(containerRef, () => setOpen(false));
  useEscapeKey(() => setOpen(false), open);

  const enabled = isCloud && !!token;
  const { invitations, totalCount, isLoading } = useUserInvitations({
    status: "pending",
    perPage: 10,
    enabled,
  });
  const acceptInvite = useAcceptInvite();
  const declineInvite = useDeclineInvite();
  const switchNamespace = useSwitchNamespace();

  if (!enabled) return null;

  // Badge and header reflect the *total* pending count, not the page we
  // happen to have fetched — otherwise the badge caps at perPage and lies
  // to users with lots of invitations.
  const count = totalCount;
  const hasMore = invitations.length < totalCount;

  const handleAccept = async () => {
    if (!acceptTarget) return;
    const tenant = acceptTarget.namespace.tenant_id;
    setAcceptError(null);
    try {
      await acceptInvite.mutateAsync({ path: { tenant } });
      // useSwitchNamespace mints a fresh namespace-scoped token, stores
      // { token, tenant, role } via setSession, and hard-navigates to
      // /dashboard so NamespaceGuard re-initializes with a clean slate.
      await switchNamespace.mutateAsync({
        tenantId: tenant,
        redirectTo: "/dashboard",
      });
      setAcceptTarget(null);
      setOpen(false);
    } catch {
      setAcceptError("Failed to accept the invitation. Please try again.");
    }
  };

  const handleDecline = async () => {
    if (!declineTarget) return;
    setDeclineError(null);
    try {
      await declineInvite.mutateAsync({
        path: { tenant: declineTarget.namespace.tenant_id },
      });
      setDeclineTarget(null);
    } catch {
      setDeclineError("Failed to decline the invitation. Please try again.");
    }
  };

  return (
    <>
      <div ref={containerRef} className="relative">
        <button
          type="button"
          onClick={() => setOpen(!open)}
          className="relative flex items-center justify-center h-8 w-8 rounded-lg border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
          aria-label={`Pending invitations${count > 0 ? ` (${count})` : ""}`}
          aria-haspopup="menu"
          aria-expanded={open}
        >
          <EnvelopeIcon
            className="w-4 h-4 text-text-secondary"
            strokeWidth={1.8}
          />
          {count > 0 && (
            <span
              aria-hidden="true"
              className="absolute -top-0.5 -right-0.5 min-w-[14px] h-[14px] px-1 rounded-full bg-accent-red text-white text-[9px] font-mono font-bold flex items-center justify-center border border-surface"
            >
              {count > 9 ? "9+" : count}
            </span>
          )}
        </button>

        {open && (
          <div
            role="menu"
            aria-label="Pending invitations"
            className="absolute top-full right-0 mt-1.5 w-80 bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down"
          >
            <div className="flex items-center justify-between px-3.5 py-3 border-b border-border">
              <div>
                <p className="text-sm font-semibold text-text-primary">
                  Invitations
                </p>
                <p className="text-2xs text-text-muted mt-0.5">
                  {count > 0
                    ? `${count} pending invitation${count !== 1 ? "s" : ""}`
                    : "No pending invitations"}
                </p>
              </div>
              <span className="inline-flex items-center justify-center w-7 h-7 rounded-md bg-primary/10 border border-primary/20">
                <InboxIcon
                  className="w-3.5 h-3.5 text-primary"
                  strokeWidth={2}
                />
              </span>
            </div>

            <div className="max-h-[360px] overflow-y-auto">
              {isLoading ? (
                <div
                  className="flex items-center justify-center gap-2 py-8"
                  role="status"
                  aria-live="polite"
                >
                  <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                  <span className="text-2xs font-mono text-text-muted">
                    Loading...
                  </span>
                </div>
              ) : count === 0 ? (
                <div className="px-3.5 py-10 text-center">
                  <InboxIcon
                    className="w-8 h-8 text-text-muted/30 mx-auto mb-2"
                    strokeWidth={1}
                  />
                  <p className="text-xs text-text-muted">
                    You&apos;re all caught up
                  </p>
                  <p className="text-2xs text-text-muted/60 mt-0.5">
                    New invitations will appear here
                  </p>
                </div>
              ) : (
                <ul className="divide-y divide-border/60">
                  {invitations.map((inv) => (
                    <InvitationCard
                      key={`${inv.namespace.tenant_id}-${inv.user.id}`}
                      invitation={inv}
                      onAccept={setAcceptTarget}
                      onDecline={setDeclineTarget}
                    />
                  ))}
                </ul>
              )}
            </div>

            {hasMore && (
              <Link
                to="/team"
                onClick={() => setOpen(false)}
                className="flex items-center justify-center gap-1.5 px-3.5 py-2.5 border-t border-border text-2xs font-semibold text-primary hover:bg-primary/5 transition-colors"
              >
                View all {totalCount} invitations
                <ArrowRightIcon className="w-3 h-3" strokeWidth={2.5} />
              </Link>
            )}
          </div>
        )}
      </div>

      <ConfirmDialog
        open={!!acceptTarget}
        onClose={() => {
          setAcceptTarget(null);
          setAcceptError(null);
        }}
        onConfirm={handleAccept}
        title="Accept Invitation"
        description={
          acceptTarget ? (
            <>
              Accept the invitation to join{" "}
              <span className="font-medium text-text-primary">
                {acceptTarget.namespace.name}
              </span>
              ? You will be switched to this namespace after accepting.
            </>
          ) : null
        }
        confirmLabel="Accept"
        variant="primary"
        errorMessage={acceptError}
      />

      <ConfirmDialog
        open={!!declineTarget}
        onClose={() => {
          setDeclineTarget(null);
          setDeclineError(null);
        }}
        onConfirm={handleDecline}
        title="Decline Invitation"
        description={
          declineTarget ? (
            <>
              Decline the invitation to join{" "}
              <span className="font-medium text-text-primary">
                {declineTarget.namespace.name}
              </span>
              ? The owner can send a new invitation later.
            </>
          ) : null
        }
        confirmLabel="Decline"
        variant="danger"
        errorMessage={declineError}
      />
    </>
  );
}
