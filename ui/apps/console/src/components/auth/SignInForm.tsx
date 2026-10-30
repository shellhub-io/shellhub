import { useState, type FormEvent, type ReactNode } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useLockout } from "@/hooks/useLockout";
import { landingAfterSignIn } from "@/utils/navigation";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { loginSchema, type LoginFormValues } from "@/pages/setup/schemas";
import MfaCodeForm from "@/components/mfa/MfaCodeForm";
import AuthActions, { type AuthLink } from "@/components/auth/AuthActions";
import LockoutCallouts from "@/components/auth/LockoutCallouts";

interface SignInFormProps {
  redirect: string;
  submitLabel?: string;
  onSignedIn?: () => void;
  links?: AuthLink[];
  secondaryAction?: ReactNode;
}

/**
 * The username and password form, with everything a sign-in can answer with: wrong credentials,
 * an unconfirmed account, one waiting for approval, a lockout with its countdown, and a second
 * factor. Without `onSignedIn` it is a page's form: success goes to `redirect`, or to a pending
 * device code when that is the dashboard, and a second factor continues on the MFA page, which
 * comes back to `redirect`. With `onSignedIn` the screen carries on in place, so the second factor
 * is asked for right here and `onSignedIn` runs once the session is full. `links` are the ways out
 * of the screen, shown under the button of whichever step is on screen. `secondaryAction` sits
 * under the sign-in button on the first step only.
 */
export default function SignInForm({
  redirect,
  submitLabel = "Sign In",
  onSignedIn,
  links = [],
  secondaryAction,
}: SignInFormProps) {
  const navigate = useNavigate();
  const { login, loading } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const [secondFactor, setSecondFactor] = useState(false);
  const { lockFrom, clearLockout, countdownDisplay, lockoutExpired } =
    useLockout();

  const { control, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    mode: "onTouched",
    defaultValues: { username: "", password: "" },
  });

  const onSubmit = async (values: LoginFormValues) => {
    setError(null);
    clearLockout();
    try {
      await login(values.username, values.password);

      if (useAuthStore.getState().mfaToken) {
        if (onSignedIn) {
          setSecondFactor(true);
          return;
        }
        const mfaPath =
          redirect !== "/dashboard"
            ? `/mfa-login?redirect=${encodeURIComponent(redirect)}`
            : "/mfa-login";
        void navigate(mfaPath);
        return;
      }

      if (onSignedIn) {
        onSignedIn();
      } else {
        void navigate(landingAfterSignIn(redirect));
      }
    } catch (err) {
      if (!isSdkError(err)) {
        setError("Something went wrong. Please try again later.");
        return;
      }

      switch (err.status) {
        case 401:
          setError(
            "Invalid login credentials. Your password is incorrect or this account doesn't exist.",
          );
          break;
        case 403:
          void navigate(
            `/confirm-account?identifier=${encodeURIComponent(values.username)}`,
            { state: { sendCode: true } },
          );
          break;
        case 423:
          setError(
            "Your account is waiting for an administrator to approve it. You'll be able to sign in once it's approved.",
          );
          break;
        case 429: {
          lockFrom(err.headers);
          setError(
            "Too many failed login attempts. Please wait before trying again.",
          );
          break;
        }
        default:
          setError("Something went wrong on our end. Please try again later.");
      }
    }
  };

  const handleFormSubmit = (e: FormEvent) => {
    void handleSubmit(onSubmit)(e);
  };

  if (secondFactor && onSignedIn) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-text-secondary">
          Enter the 6-digit code from your authenticator app to complete sign
          in.
        </p>
        <MfaCodeForm
          onVerified={onSignedIn}
          submitLabel={submitLabel}
          links={[
            { label: "Use a recovery code", to: "/mfa-recover" },
            ...links,
          ]}
        />
      </div>
    );
  }

  return (
    <form onSubmit={handleFormSubmit} className="space-y-5">
      <LockoutCallouts
        error={error}
        countdownDisplay={countdownDisplay}
        lockoutExpired={lockoutExpired}
        expiredMessage="Your timeout has finished. Please try to log back in."
      />

      <FormInputField<LoginFormValues>
        id="username"
        label="Username"
        name="username"
        control={control}
        placeholder="username"
        autoComplete="username"
      />

      <FormPasswordField<LoginFormValues>
        id="password"
        label="Password"
        name="password"
        control={control}
        placeholder="password"
        autoComplete="current-password"
      />

      <AuthActions
        primary={
          <Button
            variant="primary"
            size="lg"
            fullWidth
            type="submit"
            loading={loading}
            disabled={!formState.isValid || loading}
          >
            {loading ? "Authenticating..." : submitLabel}
          </Button>
        }
        secondary={secondaryAction}
        links={links}
      />
    </form>
  );
}
