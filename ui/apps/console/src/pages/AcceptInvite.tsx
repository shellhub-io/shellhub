import { ComponentType, SVGProps, useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useForm } from "react-hook-form";
import {
  CheckCircleIcon,
  ArrowRightIcon,
  UserPlusIcon,
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
import { inviteResolver, type InviteFormValues } from "./setup/inviteResolver";
import ScreenIntro from "@/components/layout/ScreenIntro";
import AuthActions from "@/components/auth/AuthActions";

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

/**
 * The page an invitation link lands on. It works signed out as well as in: an invitation may
 * arrive before the account exists, so this leads to sign-up and back rather than refusing.
 */
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

  const messages: Partial<Record<Branch, InvitationMessageProps>> = {
    "missing-params": {
      title: "Invalid Invitation",
      description:
        "This invitation link is missing its code. Please use the link from the original email.",
    },
    error: {
      title: "Invitation Unavailable",
      description:
        "This invitation is invalid or has expired. Please ask the sender for a new one.",
    },
    "wrong-user": {
      title: "Different Account Signed In",
      description: (
        <>
          <span>You're signed in as </span>
          <span className="font-medium text-text-primary font-mono">
            {authEmail ?? "another account"}
          </span>
          <span>
            . Sign out and use the account this invitation was sent to.
          </span>
        </>
      ),
      action: { label: "Sign Out", onClick: handleSignOut },
    },
    "sign-up": {
      title: "You've been invited",
      descriptionId: "invite-email-hint",
      description: (
        <>
          <span>Set up your account to join. You're joining as </span>
          <span className="font-medium text-text-primary font-mono">
            {inviteEmail || "your email"}
          </span>
          <span>.</span>
        </>
      ),
      children: (
        <>
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
        </>
      ),
    },
    "pending-approval": {
      title: "Waiting for Approval",
      description:
        "Your account was created and is waiting for an administrator to approve it. You'll be able to sign in once it's approved.",
    },
    joined: {
      title: "You're in",
      description: (
        <>
          <span>Your account is now a member of the namespace</span>
          {inviteEmail ? (
            <>
              <span> as </span>
              <span className="font-medium text-text-primary font-mono">
                {inviteEmail}
              </span>
            </>
          ) : null}
          <span>.</span>
        </>
      ),
      action: {
        label: "Go to Dashboard",
        onClick: () => void handleEnterNamespace(),
        loading: switchNamespace.isPending,
      },
      children: (
        <>
          <ErrorCallout message={error} />
          {switchNamespace.isPending && (
            <p className="sr-only" role="status">
              Switching to namespace…
            </p>
          )}
        </>
      ),
    },
    accept: {
      title: "Namespace Invitation",
      description:
        "Accepting this invitation will add you to the namespace. You will be automatically switched to it after accepting.",
      action: {
        label: "Accept",
        onClick: () => setShowConfirm(true),
        icon: CheckCircleIcon,
      },
    },
  };

  const message = messages[branch];

  return (
    <>
      {branch === "loading" && (
        <div
          className="flex items-center gap-3 text-sm text-text-muted"
          role="status"
          aria-live="polite"
        >
          <Spinner />
          Checking invitation...
        </div>
      )}

      {message && <InvitationMessage {...message} />}

      <ConfirmDialog
        open={showConfirm}
        onClose={() => {
          setShowConfirm(false);
          setError("");
        }}
        onConfirm={handleAccept}
        icon={<UserPlusIcon />}
        title="Accept invitation"
        description="You join the namespace and switch to it right away."
        confirmLabel="Accept invitation"
        variant="primary"
        errorMessage={error || null}
      />
    </>
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

interface InvitationActionProps {
  label: string;
  onClick: () => void;
  icon?: ComponentType<SVGProps<SVGSVGElement>>;
  loading?: boolean;
}

interface InvitationMessageProps {
  title: string;
  description: React.ReactNode;
  descriptionId?: string;
  action?: InvitationActionProps;
  children?: React.ReactNode;
}

function InvitationMessage({
  title,
  description,
  descriptionId,
  action,
  children,
}: InvitationMessageProps) {
  const backToSignIn = [{ label: "Back to sign in", to: "/login" }];

  return (
    <div>
      <ScreenIntro
        eyebrow="Invitation"
        title={title}
        lead={<span id={descriptionId}>{description}</span>}
      />
      {children}
      <AuthActions
        primary={action && <InvitationAction {...action} />}
        links={action ? undefined : backToSignIn}
      />
    </div>
  );
}

function InvitationAction({
  label,
  onClick,
  icon: ActionIcon,
  loading,
}: InvitationActionProps) {
  const iconProps = ActionIcon
    ? { icon: <ActionIcon className="w-4 h-4" strokeWidth={2} /> }
    : { iconRight: <ArrowRightIcon className="w-4 h-4" strokeWidth={2} /> };

  return (
    <Button
      onClick={onClick}
      size="lg"
      fullWidth
      loading={loading}
      {...iconProps}
    >
      {label}
    </Button>
  );
}
