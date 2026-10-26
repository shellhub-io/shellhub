import type { FormEvent } from "react";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useAuthStore } from "@/stores/authStore";
import { useLockout } from "@/hooks/useLockout";
import { useOtpInput } from "@/hooks/useOtpInput";
import OtpCells from "@/components/mfa/OtpCells";
import AuthActions, { type AuthLink } from "@/components/auth/AuthActions";
import LockoutCallouts from "@/components/auth/LockoutCallouts";

interface MfaCodeFormProps {
  onVerified: () => void;
  submitLabel?: string;
  links: AuthLink[];
}

/**
 * The second step of a sign-in: the six cells for the authenticator code, exchanged for a full
 * session. A wrong code clears the cells and leaves the partial session in place, so the user
 * can try again without signing in from scratch. Too many wrong codes lock the form out, with a
 * countdown to when the user may try again, still without signing in from scratch. `onVerified`
 * runs once the session is full.
 */
export default function MfaCodeForm({
  onVerified,
  submitLabel = "Verify",
  links,
}: MfaCodeFormProps) {
  const otp = useOtpInput(6);
  const { loginWithMfa, loading, error } = useAuthStore();
  const { lockFrom, clearLockout, countdownDisplay, lockoutExpired } =
    useLockout();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otp.isComplete) return;

    clearLockout();
    try {
      await loginWithMfa(otp.getValue());
      onVerified();
    } catch (err) {
      if (isSdkError(err) && err.status === 429) {
        lockFrom(err.headers);
      }
      otp.reset();
    }
  };

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="space-y-6">
      <LockoutCallouts
        error={error}
        countdownDisplay={countdownDisplay}
        lockoutExpired={lockoutExpired}
        expiredMessage="Your timeout has finished. Please enter a new code."
      />

      <OtpCells otp={otp} label="Verification Code" size="lg" numeric />

      <AuthActions
        primary={
          <Button
            variant="primary"
            size="lg"
            fullWidth
            type="submit"
            loading={loading}
            disabled={loading || !otp.isComplete}
          >
            {loading ? "Verifying..." : submitLabel}
          </Button>
        }
        links={links}
      />
    </form>
  );
}
