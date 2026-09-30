import { useEffect } from "react";
import { Navigate, useSearchParams } from "react-router-dom";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useSignUpStore } from "../stores/signUpStore";
import { useResendEmail } from "../hooks/useResendEmail";
import AuthActions from "@/components/auth/AuthActions";
import ScreenIntro from "@/components/layout/ScreenIntro";

/**
 * The page telling a new account to check its mail, with the resend.
 */
export default function ConfirmAccount() {
  const [searchParams] = useSearchParams();
  const resetErrors = useSignUpStore((s) => s.resetResendError);

  const username = searchParams.get("username") ?? "";

  useEffect(() => {
    resetErrors();
  }, [resetErrors]);

  const {
    handleResend,
    resendLoading,
    resendError,
    resendSuccess,
    resendCooldown,
  } = useResendEmail(username);

  if (!username) return <Navigate to="/login" replace />;

  return (
    <>
      <ScreenIntro
        eyebrow="Account"
        title="Account Activation Required"
        lead="Thank you for registering an account on ShellHub. An email was sent with a confirmation link. You need to click on the link to activate your account. If you haven't received the email, click on the Resend Email button."
      />

      <div className="flex flex-col gap-3 mb-6 empty:hidden">
        {resendSuccess && (
          <Callout variant="success">
            Confirmation email sent successfully.
          </Callout>
        )}
        {resendError && <Callout variant="error">{resendError}</Callout>}
      </div>

      <AuthActions
        primary={
          <Button
            size="lg"
            fullWidth
            loading={resendLoading}
            disabled={resendLoading || resendCooldown > 0}
            onClick={() => void handleResend()}
          >
            {resendLoading
              ? "Sending..."
              : resendCooldown > 0
                ? `Resend Email (${resendCooldown}s)`
                : "Resend Email"}
          </Button>
        }
        links={[{ label: "Back to sign in", to: "/login" }]}
      />
    </>
  );
}
