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
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { useAcceptInvite } from "@/hooks/useInvitationMutations";
import { useResolveInvitation } from "@/hooks/useInvitations";
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
  | "loading"
  | "missing-params"
  | "error"
  | "wrong-user"
  | "sign-up"
  | "pending-approval"
  | "joined"
  | "accept";

type PostAction =
  { kind: "pending-approval" } | { kind: "joined"; token?: string };

export default function AcceptInvite() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const authUserId = useAuthStore((s) => s.userId);
  const authEmail = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);
  const loginWithToken = useAuthStore((s) => s.loginWithToken);

  const invite = searchParams.get("invite") ?? "";

  const acceptInvite = useAcceptInvite();
  const switchNamespace = useSwitchNamespace();

  const signUp = useSignUpStore((s) => s.signUp);
  const signUpLoading = useSignUpStore((s) => s.signUpLoading);
  const signUpError = useSignUpStore((s) => s.signUpError);

  const { resolved, isLoading, isError } = useResolveInvitation(invite);

  const tenant = resolved?.tenantId ?? "";
  const inviteEmail = resolved?.email ?? "";

  const [postAction, setPostAction] = useState<PostAction | null>(null);
  const [showConfirm, setShowConfirm] = useState(false);
  const [error, setError] = useState("");

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

  const needsLogin =
    !authToken &&
    !postAction &&
    !!resolved &&
    (resolved.status === "not-confirmed" || resolved.status === "confirmed");

  useEffect(() => {
    if (!needsLogin) return;
    const redirectTarget = `/accept-invite?invite=${encodeURIComponent(invite)}`;
    void navigate(`/login?redirect=${encodeURIComponent(redirectTarget)}`);
  }, [needsLogin, invite, navigate]);

  const branch: Branch = (() => {
    if (postAction) return postAction.kind;
    if (!invite) return "missing-params";
    if (isLoading || needsLogin) return "loading";
    if (isError || !resolved) return "error";

    if (authToken) {
      return authUserId === resolved.userId ? "accept" : "wrong-user";
    }

    if (resolved.status === "invited") return "sign-up";

    return "error";
  })();

  const handleSignUp = async (values: InviteFormValues) => {
    setError("");

    const token = await signUp({
      name: values.name,
      username: values.username,
      email: inviteEmail,
      password: values.password,
      email_marketing: false,
      sig: invite,
    });

    const { signUpError: err, signUpServerFields } = useSignUpStore.getState();
    if (err) return;
    if (signUpServerFields.length > 0) {
      setError(
        "That username or email is already in use. Try a different username.",
      );
      return;
    }

    if (token) {
      setPostAction({ kind: "joined", token });
      return;
    }

    setPostAction({ kind: "pending-approval" });
  };

  const handleAccept = async () => {
    if (!tenant || !authToken) return;
    setError("");
    try {
      await acceptInvite.mutateAsync({ path: { tenant } });
      setShowConfirm(false);
      setPostAction({ kind: "joined" });
    } catch {
      setError("Failed to accept the invitation. Please try again.");
    }
  };

  const handleEnterNamespace = async () => {
    setError("");
    try {
      if (postAction?.kind === "joined" && postAction.token) {
        await loginWithToken(postAction.token);
      }

      await switchNamespace.mutateAsync({
        tenantId: tenant,
        redirectTo: "/dashboard",
      });
    } catch {
      setError("Couldn't open the namespace. Please try again.");
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
        {branch === "loading" && (
          <div
            className="flex flex-col items-center gap-3 py-6"
            role="status"
            aria-live="polite"
          >
            <Spinner size="2xl" />
            <p className="text-sm text-text-muted">Checking invitation...</p>
          </div>
        )}

        {branch === "missing-params" && (
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

        {branch === "error" && (
          <InvitationMessage
            tone="error"
            icon={
              <ExclamationTriangleIcon
                className="w-7 h-7 text-accent-red"
                strokeWidth={1.5}
              />
            }
            title="Invitation Unavailable"
            description="This invitation is invalid or has expired. Please ask the sender for a new one."
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

        {branch === "wrong-user" && (
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

        {branch === "sign-up" && (
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
              <p
                id="invite-email-hint"
                className="text-sm text-text-secondary leading-relaxed"
              >
                Set up your account to join. You&apos;re joining as{" "}
                <span className="font-medium text-text-primary font-mono">
                  {inviteEmail || "your email"}
                </span>
                .
              </p>
            </div>

            <ErrorCallout message={signUpError} />
            <ErrorCallout message={error} />

            <form
              onSubmit={(e) => void handleSubmit(handleSignUp)(e)}
              className="space-y-4"
              aria-label="Complete your account"
              aria-describedby="invite-email-hint"
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

        {branch === "pending-approval" && (
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

        {branch === "joined" && (
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
            <ErrorCallout message={error} />
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
            {switchNamespace.isPending && (
              <p className="sr-only" role="status">
                Switching to namespace…
              </p>
            )}
          </div>
        )}

        {branch === "accept" && (
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
                onClick={() => setShowConfirm(true)}
              >
                Accept
              </Button>
            </div>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={showConfirm}
        onClose={() => {
          setShowConfirm(false);
          setError("");
        }}
        onConfirm={handleAccept}
        title="Accept Invitation"
        description="You will be added to the namespace and switched to it immediately."
        confirmLabel="Accept"
        variant="primary"
        errorMessage={error || null}
      />
    </div>
  );
}

function ErrorCallout({ message }: { message: string | null }) {
  if (!message) return null;
  return (
    <Callout variant="error" className="mb-4">
      {message}
    </Callout>
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
        className={cn(
          "inline-flex items-center justify-center w-14 h-14 rounded-full border mb-5",
          ringClass,
        )}
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
