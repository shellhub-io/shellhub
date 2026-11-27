import type { FormEvent } from "react";
import { useSearchParams, Link } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { mfaResetCodesRejectedMessage } from "@/stores/authErrors";
import { useCompleteMfaReset } from "@/hooks/useUserRequests";
import { useOtpInput } from "@/hooks/useOtpInput";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import OtpCells from "@/components/mfa/OtpCells";

/**
 * The last step of an MFA reset, reached from the emailed link. The reset signs no one in, so it
 * ends on the sign-in page.
 */
export default function MfaResetComplete() {
  const [searchParams] = useSearchParams();
  const reset = useCompleteMfaReset();
  const loading = reset.isPending;
  const error = reset.error ? mfaResetCodesRejectedMessage : null;

  const otpMain = useOtpInput(5, true);
  const otpRecovery = useOtpInput(5, true);

  const userId = searchParams.get("id");

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!userId || !otpMain.isComplete || !otpRecovery.isComplete) return;

    reset.mutate(
      {
        path: { "user-id": userId },
        body: {
          main_email_code: otpMain.getValue(),
          recovery_email_code: otpRecovery.getValue(),
        },
      },
      {
        onError: () => {
          otpMain.reset();
          otpRecovery.reset();
        },
      },
    );
  };

  if (!userId) {
    return (
      <>
        <ScreenIntro
          eyebrow="Two-factor"
          title="Invalid Reset Link"
          lead="This reset link is invalid or has expired. Please request a new one."
        />
        <AuthActions
          primary={
            <Button as={Link} to="/mfa-reset-request" size="lg" fullWidth>
              Request a new reset link
            </Button>
          }
          links={[{ label: "Back to sign in", to: "/login" }]}
        />
      </>
    );
  }

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Enter Verification Codes"
        lead="Enter the codes from both emails to complete the MFA reset. Both are required to prove ownership of your account's email addresses, and they expire after 24 hours."
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && <Callout variant="error">{error}</Callout>}

        <OtpCells
          otp={otpMain}
          label="Main Email Code"
          hint="Code sent to your main email address"
        />

        <OtpCells
          otp={otpRecovery}
          label="Recovery Email Code"
          hint="Code sent to your recovery email address"
        />

        <AuthActions
          primary={
            <Button
              variant="primary"
              size="lg"
              fullWidth
              type="submit"
              loading={loading}
              disabled={
                !otpMain.isComplete || !otpRecovery.isComplete || loading
              }
            >
              {loading ? "Verifying..." : "Reset MFA"}
            </Button>
          }
          links={[{ label: "Back to sign in", to: "/login" }]}
        />
      </form>
    </>
  );
}
