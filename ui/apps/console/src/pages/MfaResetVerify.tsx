import { FormEvent, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useMfaResetStore } from "../stores/mfaResetStore";
import { useOtpInput } from "../hooks/useOtpInput";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";
import OtpCells from "@/components/mfa/OtpCells";

/**
 * Verifies the code from an MFA reset mail. The request expires, so this counts down rather than
 * waiting indefinitely.
 */
export default function MfaResetVerify() {
  const {
    completeMfaReset,
    mfaResetToken,
    mfaResetIdentifier,
    loading,
    error,
  } = useMfaResetStore();
  const navigate = useNavigate();

  const otpMain = useOtpInput(5, true);
  const otpRecovery = useOtpInput(5, true);

  useEffect(() => {
    useMfaResetStore.setState({ error: null });
  }, []);

  useEffect(() => {
    if (!mfaResetToken) {
      void navigate("/mfa-recover");
    }
  }, [mfaResetToken, navigate]);

  if (!mfaResetToken) {
    return null;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otpMain.isComplete || !otpRecovery.isComplete) return;

    try {
      await completeMfaReset(otpMain.getValue(), otpRecovery.getValue());
      void navigate("/dashboard");
    } catch {
      otpMain.reset();
      otpRecovery.reset();
    }
  };

  return (
    <>
      <ScreenIntro
        eyebrow="Two-factor"
        title="Enter Verification Codes"
        lead={`Check both email addresses for ${mfaResetIdentifier} and enter the codes below. Both are required, and they expire after 24 hours.`}
      />

      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-6">
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
