import { ComponentType, SVGProps, useEffect, useState } from "react";
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
import LoginLayoutCard from "@/components/layout/LoginLayoutCard";

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

  const messages: Partial<Record<Branch, InvitationMessageProps>> = {
    "missing-params": {
      tone: "error",
      icon: XCircleIcon,
      title: "Invalid Invitation",
      description:
        "This invitation link is missing its code. Please use the link from the original email.",
      action: { label: "Back to Login", to: "/login" },
    },
    error: {
      tone: "error",
      icon: ExclamationTriangleIcon,
      title: "Invitation Unavailable",
      description:
        "This invitation is invalid or has expired. Please ask the sender for a new one.",
      action: { label: "Back to Login", to: "/login" },
    },
    "wrong-user": {
      tone: "warning",
      icon: UserCircleIcon,
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
      tone: "primary",
      icon: EnvelopeOpenIcon,
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
      tone: "warning",
      icon: ClockIcon,
      title: "Waiting for Approval",
      description:
        "Your account was created and is waiting for an administrator to approve it. You'll be able to sign in once it's approved.",
      action: { label: "Back to Login", to: "/login" },
    },
    joined: {
      tone: "success",
      icon: CheckCircleIcon,
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
      tone: "primary",
      icon: EnvelopeOpenIcon,
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
      <LoginLayoutCard>
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

        {message && <InvitationMessage {...message} />}
      </LoginLayoutCard>

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

type Tone = "error" | "warning" | "primary" | "success";

const toneStyles: Record<Tone, { ring: string; iconColor: string }> = {
  error: {
    ring: "bg-accent-red/10 border-accent-red/20",
    iconColor: "text-accent-red",
  },
  warning: {
    ring: "bg-accent-yellow/10 border-accent-yellow/20",
    iconColor: "text-accent-yellow",
  },
  primary: {
    ring: "bg-primary/10 border-primary/20",
    iconColor: "text-primary",
  },
  success: {
    ring: "bg-accent-green/10 border-accent-green/20",
    iconColor: "text-accent-green",
  },
};

interface InvitationActionProps {
  label: string;
  to?: string;
  onClick?: () => void;
  icon?: ComponentType<SVGProps<SVGSVGElement>>;
  loading?: boolean;
}

interface InvitationMessageProps {
  tone: Tone;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  title: string;
  description: React.ReactNode;
  descriptionId?: string;
  action?: InvitationActionProps;
  children?: React.ReactNode;
}

function InvitationMessage({
  tone,
  icon: Icon,
  title,
  description,
  descriptionId,
  action,
  children,
}: InvitationMessageProps) {
  const { ring, iconColor } = toneStyles[tone];

  return (
    <div>
      <div className="text-center">
        <div
          className={cn(
            "inline-flex items-center justify-center w-14 h-14 rounded-full border mb-5",
            ring,
          )}
        >
          <Icon className={cn("w-7 h-7", iconColor)} strokeWidth={1.5} />
        </div>
        <h1 className="text-lg font-semibold text-text-primary mb-3">
          {title}
        </h1>
        <p
          id={descriptionId}
          className="text-sm text-text-secondary leading-relaxed mb-6"
        >
          {description}
        </p>
      </div>
      {children}
      {action && <InvitationAction {...action} />}
    </div>
  );
}

function InvitationAction({
  label,
  to,
  onClick,
  icon: ActionIcon,
  loading,
}: InvitationActionProps) {
  const iconProps = ActionIcon
    ? { icon: <ActionIcon className="w-4 h-4" strokeWidth={2} /> }
    : { iconRight: <ArrowRightIcon className="w-4 h-4" strokeWidth={2} /> };

  return (
    <div className="text-center">
      {to ? (
        <Button as={Link} to={to} loading={loading} {...iconProps}>
          {label}
        </Button>
      ) : (
        <Button onClick={onClick} loading={loading} {...iconProps}>
          {label}
        </Button>
      )}
    </div>
  );
}
