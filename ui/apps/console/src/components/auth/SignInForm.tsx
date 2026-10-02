import { useState, FormEvent } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import {
  lockoutEndFrom,
  useLockoutCountdown,
} from "@/hooks/useLockoutCountdown";
import { landingAfterSignIn } from "@/utils/navigation";
import {
  FormInputField,
  FormPasswordField,
} from "@/components/common/fields/rhf";
import { loginResolver } from "@/pages/setup/loginResolver";
import MfaCodeForm from "@/components/mfa/MfaCodeForm";
import AuthActions, { type AuthLink } from "@/components/auth/AuthActions";
import type { LoginFormValues } from "@/pages/setup/loginResolver";

interface SignInFormProps {
  redirect: string;
  submitLabel?: string;
  onSignedIn?: () => void;
  links?: AuthLink[];
}

/**
 * The username and password form, with everything a sign-in can answer with: wrong credentials,
 * an unconfirmed account, one waiting for approval, a lockout with its countdown, and a second
 * factor. Without `onSignedIn` it is a page's form: success goes to `redirect`, or to a pending
 * device code when that is the dashboard, and a second factor continues on the MFA page, which
 * comes back to `redirect`. With `onSignedIn` the screen carries on in place, so the second factor
 * is asked for right here and `onSignedIn` runs once the session is full. `links` are the ways out
 * of the screen, shown under the button of whichever step is on screen.
 */
export default function SignInForm({
  redirect,
  submitLabel = "Sign In",
  onSignedIn,
  links = [],
}: SignInFormProps) {
  const navigate = useNavigate();
  const { login, loading } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const [secondFactor, setSecondFactor] = useState(false);
  const [lockoutEndEpoch, setLockoutEndEpoch] = useState<number | null>(null);
  const { display: countdownDisplay, expired: lockoutExpired } =
    useLockoutCountdown(lockoutEndEpoch);

  const { control, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: loginResolver,
    mode: "onTouched",
    defaultValues: { username: "", password: "" },
  });

  const onSubmit = async (values: LoginFormValues) => {
    setError(null);
    setLockoutEndEpoch(null);
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
            `/confirm-account?username=${encodeURIComponent(values.username)}`,
          );
          break;
        case 423:
          setError(
            "Your account is waiting for an administrator to approve it. You'll be able to sign in once it's approved.",
          );
          break;
        case 429: {
          setLockoutEndEpoch(lockoutEndFrom(err.headers));
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
      {lockoutExpired && (
        <Callout variant="success">
          Your timeout has finished. Please try to log back in.
        </Callout>
      )}
      {error && !lockoutExpired && (
        <Callout variant="error">
          <span>
            <span>{error}</span>
            {countdownDisplay ? (
              <span className="font-semibold"> ({countdownDisplay})</span>
            ) : null}
          </span>
        </Callout>
      )}

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
        links={links}
      />
    </form>
  );
}
