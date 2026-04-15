import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import {
  EnvelopeOpenIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowRightIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import { lookupUserStatus } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import {
  useAcceptInvite,
  useDeclineInvite,
} from "@/hooks/useInvitationMutations";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";

type Branch =
  | { kind: "loading" }
  | { kind: "missing-params" }
  | { kind: "error"; message: string }
  | { kind: "wrong-user" }
  | { kind: "ready" };

export default function AcceptInvite() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const authUserId = useAuthStore((s) => s.userId);
  const authEmail = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);

  const tenant = searchParams.get("tenant-id") ?? "";
  const userId = searchParams.get("user-id") ?? "";
  const sig = searchParams.get("sig") ?? "";
  const email = searchParams.get("email") ?? "";

  const acceptInvite = useAcceptInvite();
  const declineInvite = useDeclineInvite();
  const switchNamespace = useSwitchNamespace();

  const [branch, setBranch] = useState<Branch>({ kind: "loading" });
  const [confirmKind, setConfirmKind] = useState<"accept" | "decline" | null>(
    null,
  );
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function resolve() {
      if (!tenant || !userId || !sig) {
        if (!cancelled) setBranch({ kind: "missing-params" });
        return;
      }

      if (authToken) {
        if (authUserId === userId) {
          if (!cancelled) setBranch({ kind: "ready" });
        } else if (!cancelled) {
          setBranch({ kind: "wrong-user" });
        }
        return;
      }

      try {
        const { data } = await lookupUserStatus({
          path: { tenant, id: userId },
          query: { sig },
          throwOnError: true,
        });
        if (cancelled) return;

        const status = data.status;

        if (status === "invited") {
          // Forward every invitation param so the post-signup redirect back
          // to /accept-invite lands with the full context intact. Dropping
          // tenant-id or user-id here strands the user on the "Invalid
          // Invitation" card after they finish creating their account.
          const signUpQuery = new URLSearchParams();
          if (email) signUpQuery.set("email", email);
          if (sig) signUpQuery.set("sig", sig);
          if (tenant) signUpQuery.set("tenant-id", tenant);
          if (userId) signUpQuery.set("user-id", userId);
          void navigate(`/sign-up?${signUpQuery.toString()}`);
          return;
        }

        if (status === "not-confirmed" || status === "confirmed") {
          const redirectTarget = `/accept-invite?${searchParams.toString()}`;
          void navigate(
            `/login?redirect=${encodeURIComponent(redirectTarget)}`,
          );
          return;
        }

        if (!cancelled) {
          setBranch({
            kind: "error",
            message: "We couldn't verify this invitation. Please try again.",
          });
        }
      } catch {
        if (!cancelled) {
          setBranch({
            kind: "error",
            message:
              "This invitation is invalid or has expired. Please ask the sender for a new one.",
          });
        }
      }
    }

    void resolve();
    return () => {
      cancelled = true;
    };
  }, [
    tenant,
    userId,
    sig,
    email,
    authToken,
    authUserId,
    searchParams,
    navigate,
  ]);

  const handleAccept = async () => {
    if (!tenant || !authToken) return;
    setActionError("");
    try {
      await acceptInvite.mutateAsync({ path: { tenant } });
      // switchNamespace mints a fresh namespace-scoped token, stores
      // { token, tenant, role } via setSession, and hard-navigates so
      // NamespaceGuard re-initializes with a clean slate. We intentionally
      // don't call navigate("/dashboard") below — useSwitchNamespace owns
      // the redirect.
      await switchNamespace.mutateAsync({
        tenantId: tenant,
        redirectTo: "/dashboard",
      });
      setConfirmKind(null);
    } catch {
      setActionError("Failed to accept the invitation. Please try again.");
    }
  };

  const handleDecline = async () => {
    if (!tenant) return;
    setActionError("");
    try {
      await declineInvite.mutateAsync({ path: { tenant } });
      setConfirmKind(null);
      void navigate("/dashboard");
    } catch {
      setActionError("Failed to decline the invitation. Please try again.");
    }
  };

  const handleSignOut = () => {
    logout();
    void navigate(
      `/login?redirect=${encodeURIComponent(`/accept-invite?${searchParams.toString()}`)}`,
    );
  };

  return (
    <div className="w-full max-w-md mx-auto animate-fade-in">
      <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm">
        {branch.kind === "loading" && (
          <div
            className="flex flex-col items-center gap-3 py-6"
            role="status"
            aria-live="polite"
          >
            <span className="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
            <p className="text-sm text-text-muted">Checking invitation...</p>
          </div>
        )}

        {branch.kind === "missing-params" && (
          <InvitationMessage
            tone="error"
            icon={
              <XCircleIcon
                className="w-7 h-7 text-accent-red"
                strokeWidth={1.5}
              />
            }
            title="Invalid Invitation"
            description="This invitation link is missing required parameters. Please use the link from the original email."
            action={
              <Link
                to="/login"
                className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all"
              >
                Back to Login
                <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
              </Link>
            }
          />
        )}

        {branch.kind === "error" && (
          <InvitationMessage
            tone="error"
            icon={
              <ExclamationTriangleIcon
                className="w-7 h-7 text-accent-red"
                strokeWidth={1.5}
              />
            }
            title="Invitation Unavailable"
            description={branch.message}
            action={
              <Link
                to="/login"
                className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all"
              >
                Back to Login
                <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
              </Link>
            }
          />
        )}

        {branch.kind === "wrong-user" && (
          <InvitationMessage
            tone="warning"
            icon={
              <UserCircleIcon
                className="w-7 h-7 text-accent-yellow"
                strokeWidth={1.5}
              />
            }
            title="Different Account Signed In"
            description={
              <>
                You&apos;re signed in as{" "}
                <span className="font-medium text-text-primary font-mono">
                  {authEmail ?? "another account"}
                </span>
                . Sign out and use the account this invitation was sent to.
              </>
            }
            action={
              <button
                type="button"
                onClick={handleSignOut}
                className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all"
              >
                Sign Out
                <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
              </button>
            }
          />
        )}

        {branch.kind === "ready" && (
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 border border-primary/20 mb-5">
              <EnvelopeOpenIcon
                className="w-7 h-7 text-primary"
                strokeWidth={1.5}
              />
            </div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">
              Namespace Invitation
            </h2>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              Accepting this invitation will add you to the namespace. You will
              be automatically switched to it after accepting.
            </p>
            <div className="flex items-center justify-center gap-2">
              <button
                type="button"
                onClick={() => setConfirmKind("decline")}
                className="px-5 py-2.5 text-sm font-semibold text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Decline
              </button>
              <button
                type="button"
                onClick={() => setConfirmKind("accept")}
                className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all"
              >
                <CheckCircleIcon className="w-4 h-4" strokeWidth={2} />
                Accept
              </button>
            </div>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={confirmKind === "accept"}
        onClose={() => {
          setConfirmKind(null);
          setActionError("");
        }}
        onConfirm={handleAccept}
        title="Accept Invitation"
        description="You will be added to the namespace and switched to it immediately."
        confirmLabel="Accept"
        variant="primary"
        errorMessage={confirmKind === "accept" ? actionError || null : null}
      />

      <ConfirmDialog
        open={confirmKind === "decline"}
        onClose={() => {
          setConfirmKind(null);
          setActionError("");
        }}
        onConfirm={handleDecline}
        title="Decline Invitation"
        description="The invitation will be marked as rejected. You can ask the sender to invite you again later."
        confirmLabel="Decline"
        variant="danger"
        errorMessage={confirmKind === "decline" ? actionError || null : null}
      />
    </div>
  );
}

function InvitationMessage({
  tone,
  icon,
  title,
  description,
  action,
}: {
  tone: "error" | "warning";
  icon: React.ReactNode;
  title: string;
  description: React.ReactNode;
  action: React.ReactNode;
}) {
  const ringClass =
    tone === "error"
      ? "bg-accent-red/10 border-accent-red/20"
      : "bg-accent-yellow/10 border-accent-yellow/20";
  return (
    <div className="text-center">
      <div
        className={`inline-flex items-center justify-center w-14 h-14 rounded-full border mb-5 ${ringClass}`}
      >
        {icon}
      </div>
      <h2 className="text-lg font-semibold text-text-primary mb-3">{title}</h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-6">
        {description}
      </p>
      {action}
    </div>
  );
}
