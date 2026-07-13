import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import {
  EnvelopeOpenIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowRightIcon,
  UserCircleIcon,
  ClockIcon,
} from "@heroicons/react/24/outline";
import { resolveInvitation } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { useAcceptInvite } from "@/hooks/useInvitationMutations";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { Button, Spinner, Callout } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { inviteResolver, type InviteFormValues } from "./setup/inviteResolver";

type Branch =
  | { kind: "loading" }
  | { kind: "missing-params" }
  | { kind: "error"; message: string }
  | { kind: "wrong-user" }
  | { kind: "complete" } // account doesn't exist yet: the invitee sets it up here
  | { kind: "submitted" } // completed, waiting for a superadmin's approval
  | { kind: "joined"; token?: string } // accepted/completed and live: confirm before entering
  | { kind: "ready" }; // account exists and is signed in: accept

export default function AcceptInvite() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const authUserId = useAuthStore((s) => s.userId);
  const authEmail = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);
  const setSession = useAuthStore((s) => s.setSession);

  const invite = searchParams.get("invite") ?? "";

  const acceptInvite = useAcceptInvite();
  const switchNamespace = useSwitchNamespace();

  const signUp = useSignUpStore((s) => s.signUp);
  const signUpLoading = useSignUpStore((s) => s.signUpLoading);
  const signUpError = useSignUpStore((s) => s.signUpError);

  // Resolved from the invite code (the link no longer carries these). tenant is
  // needed to accept/decline; email is shown as context while completing.
  const [tenant, setTenant] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [branch, setBranch] = useState<Branch>({ kind: "loading" });
  const [confirmKind, setConfirmKind] = useState<"accept" | null>(null);
  const [actionError, setActionError] = useState("");
  const [completeError, setCompleteError] = useState("");

  const { control, handleSubmit } = useForm<InviteFormValues>({
    resolver: inviteResolver,
    mode: "onTouched",
    defaultValues: {
      name: "",
      username: "",
      password: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    let cancelled = false;

    async function resolve() {
      if (!invite) {
        if (!cancelled) setBranch({ kind: "missing-params" });
        return;
      }

      try {
        const { data } = await resolveInvitation({
          query: { invite },
          throwOnError: true,
        });
        if (cancelled) return;

        if (data.tenant_id) setTenant(data.tenant_id);
        if (data.email) setInviteEmail(data.email);

        // Logged in: only the invited account may accept. Compare against the
        // account the code resolves to.
        if (authToken) {
          if (authUserId === data.user_id) setBranch({ kind: "ready" });
          else setBranch({ kind: "wrong-user" });
          return;
        }

        // No account yet — the invitee sets it up right here, no generic sign-up.
        if (data.status === "invited") {
          setBranch({ kind: "complete" });
          return;
        }

        // Account exists but they aren't signed in: send them to log in, then
        // back here to accept.
        if (data.status === "not-confirmed" || data.status === "confirmed") {
          const redirectTarget = `/accept-invite?invite=${encodeURIComponent(invite)}`;
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
  }, [invite, authToken, authUserId, navigate]);

  const onComplete = async (values: InviteFormValues) => {
    setCompleteError("");

    // email comes from the invite; no ToS/marketing (that's Cloud's open sign-up).
    const token = await signUp({
      name: values.name,
      username: values.username,
      email: inviteEmail,
      password: values.password,
      email_marketing: false,
      sig: invite,
    });

    const { signUpError: err, signUpServerFields: fields } =
      useSignUpStore.getState();

    if (err) return; // shown via the signUpError Callout
    if (fields.length > 0) {
      setCompleteError(
        "That username or email is already in use. Try a different username.",
      );
      return;
    }

    if (token) {
      // Confirmed account (superadmin invite / Cloud). Carry the token so entering the namespace
      // can establish the session. Calling setSession here would flip authToken and re-run the
      // resolve effect, clobbering this screen; defer it to handleEnterNamespace instead.
      setBranch({ kind: "joined", token });
      return;
    }

    // No token: the account was created but needs a superadmin's approval
    // before it can sign in (Enterprise, non-superadmin inviter).
    setBranch({ kind: "submitted" });
  };

  const handleAccept = async () => {
    if (!tenant || !authToken) return;
    setActionError("");
    try {
      await acceptInvite.mutateAsync({ path: { tenant } });
      setConfirmKind(null);
      setBranch({ kind: "joined" });
    } catch {
      setActionError("Failed to accept the invitation. Please try again.");
    }
  };

  const handleEnterNamespace = async () => {
    setActionError("");
    try {
      // A freshly-completed account isn't signed in yet: establish the session from the completion
      // token so getNamespaceToken is authenticated. (The accept flow of an existing account is
      // already signed in and carries no token.)
      if (branch.kind === "joined" && branch.token) {
        setSession({ token: branch.token, tenant });
      }

      // switchNamespace mints a fresh namespace-scoped token, stores { token, tenant, role }
      // via setSession, and hard-navigates so NamespaceGuard re-initializes cleanly.
      await switchNamespace.mutateAsync({
        tenantId: tenant,
        redirectTo: "/dashboard",
      });
    } catch {
      setActionError("Couldn't open the namespace. Please try again.");
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
            <Spinner size="2xl" />
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
            description="This invitation link is missing its code. Please use the link from the original email."
            action={
              <Button
                as={Link}
                to="/login"
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                Back to Login
              </Button>
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
              <Button
                as={Link}
                to="/login"
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                Back to Login
              </Button>
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
              <Button
                onClick={handleSignOut}
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                Sign Out
              </Button>
            }
          />
        )}

        {branch.kind === "complete" && (
          <div>
            <div className="text-center mb-6">
              <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 border border-primary/20 mb-5">
                <EnvelopeOpenIcon
                  className="w-7 h-7 text-primary"
                  strokeWidth={1.5}
                />
              </div>
              <h2 className="text-lg font-semibold text-text-primary mb-2">
                You&apos;ve been invited
              </h2>
              <p className="text-sm text-text-secondary leading-relaxed">
                Set up your account to join. You&apos;re joining as{" "}
                <span className="font-medium text-text-primary font-mono">
                  {inviteEmail || "your email"}
                </span>
                .
              </p>
            </div>

            {signUpError && (
              <Callout variant="error" className="mb-4">
                {signUpError}
              </Callout>
            )}
            {completeError && (
              <Callout variant="error" className="mb-4">
                {completeError}
              </Callout>
            )}

            <form
              onSubmit={(e) => void handleSubmit(onComplete)(e)}
              className="space-y-4"
              aria-label="Complete your account"
            >
              <FormInputField<InviteFormValues>
                id="invite-name"
                label="Name"
                name="name"
                control={control}
                placeholder="Your name"
                autoComplete="name"
              />
              <FormInputField<InviteFormValues>
                id="invite-username"
                label="Username"
                name="username"
                control={control}
                placeholder="username"
                autoComplete="username"
              />
              <FormPasswordField<InviteFormValues>
                id="invite-password"
                label="Password"
                name="password"
                control={control}
                autoComplete="new-password"
              />
              <FormPasswordField<InviteFormValues>
                id="invite-confirm-password"
                label="Confirm password"
                name="confirmPassword"
                control={control}
                autoComplete="new-password"
              />
              <Button type="submit" className="w-full" loading={signUpLoading}>
                Join Namespace
              </Button>
            </form>
          </div>
        )}

        {branch.kind === "submitted" && (
          <InvitationMessage
            tone="warning"
            icon={
              <ClockIcon
                className="w-7 h-7 text-accent-yellow"
                strokeWidth={1.5}
              />
            }
            title="Waiting for Approval"
            description="Your account was created and is waiting for an administrator to approve it. You'll be able to sign in once it's approved."
            action={
              <Button
                as={Link}
                to="/login"
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                Back to Login
              </Button>
            }
          />
        )}

        {branch.kind === "joined" && (
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-green/10 border border-accent-green/20 mb-5">
              <CheckCircleIcon
                className="w-7 h-7 text-accent-green"
                strokeWidth={1.5}
              />
            </div>
            <h2 className="text-lg font-semibold text-text-primary mb-3">
              You&apos;re in
            </h2>
            <p className="text-sm text-text-secondary leading-relaxed mb-6">
              Your account is now a member of the namespace
              {inviteEmail ? (
                <>
                  {" "}
                  as{" "}
                  <span className="font-medium text-text-primary font-mono">
                    {inviteEmail}
                  </span>
                </>
              ) : null}
              .
            </p>
            {actionError && (
              <Callout variant="error" className="mb-4">
                {actionError}
              </Callout>
            )}
            <div className="flex items-center justify-center">
              <Button
                onClick={() => void handleEnterNamespace()}
                loading={switchNamespace.isPending}
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
              >
                Go to Dashboard
              </Button>
            </div>
          </div>
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
            <div className="flex items-center justify-center">
              <Button
                icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
                onClick={() => setConfirmKind("accept")}
              >
                Accept
              </Button>
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
        className={cn("inline-flex items-center justify-center w-14 h-14 rounded-full border mb-5", ringClass)}
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
