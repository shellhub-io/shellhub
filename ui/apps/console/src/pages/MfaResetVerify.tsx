import type { FormEvent } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { mfaResetCodesRejectedMessage } from "@/stores/authErrors";
import { useCompleteMfaReset } from "@/hooks/useUserRequests";
import { useOtpInput } from "@/hooks/useOtpInput";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import OtpCells from "@/components/mfa/OtpCells";

/**
 * Verifies the codes from the two MFA reset emails. The reset request's token and the identifier
 * it was made for arrive as route state from the request page; without them the page sends the
 * user back to recovery.
 */
export default function MfaResetVerify() {
  const { token, identifier } =
    (useLocation().state as { token?: string; identifier?: string } | null) ??
    {};
  const reset = useCompleteMfaReset();
  const loading = reset.isPending;
  const error = reset.error ? mfaResetCodesRejectedMessage : null;

  const otpMain = useOtpInput(5, true);
  const otpRecovery = useOtpInput(5, true);

  if (!token || !identifier) {
    return <Navigate to="/mfa-recover" replace />;
  }

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!otpMain.isComplete || !otpRecovery.isComplete) return;

    reset.mutate(
      {
        path: { "user-id": token },
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

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Enter Verification Codes"
        lead={`Check both email addresses for ${identifier} and enter the codes below. Both are required, and they expire after 24 hours.`}
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
              {loading ? "Verifying..." : "Verify and Reset MFA"}
            </Button>
          }
          links={[{ label: "Resend codes", to: "/mfa-reset-request" }]}
        />
      </form>
    </>
  );
}
