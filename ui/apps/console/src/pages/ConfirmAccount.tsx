import { useEffect } from "react";
import { Link, Navigate, useSearchParams } from "react-router-dom";
import { EnvelopeIcon } from "@heroicons/react/24/outline";
import Alert from "@/components/common/Alert";
import { useSignUpStore } from "../stores/signUpStore";
import { useResendEmail } from "../hooks/useResendEmail";
import { Button } from "@shellhub/design-system/primitives";

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
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      <div className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center animate-slide-up">
        <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 border border-primary/20 mb-5">
          <EnvelopeIcon className="w-7 h-7 text-primary" strokeWidth={1.5} />
        </div>

        <h1 className="text-lg font-semibold text-text-primary mb-3">
          Account Activation Required
        </h1>

        <p className="text-sm text-text-secondary leading-relaxed mb-6">
          Thank you for registering an account on ShellHub. An email was sent
          with a confirmation link. You need to click on the link to activate
          your account. If you haven&apos;t received the email, click on the
          Resend Email button.
        </p>

        {resendSuccess && (
          <Alert variant="success" className="mb-4">
            Confirmation email sent successfully.
          </Alert>
        )}

        {resendError && (
          <Alert variant="error" className="mb-4">
            {resendError}
          </Alert>
        )}

        <Button
          fullWidth
          loading={resendLoading}
          disabled={resendLoading || resendCooldown > 0}
          onClick={() => void handleResend()}
          className="mb-5"
        >
          {resendLoading
            ? "Sending..."
            : resendCooldown > 0
              ? `Resend Email (${resendCooldown}s)`
              : "Resend Email"}
        </Button>

        <p className="text-xs text-text-muted">
          Back to{" "}
          <Link
            to="/login"
            className="text-primary hover:text-primary/80 font-medium transition-colors"
          >
            Login
          </Link>
        </p>
      </div>
    </div>
  );
}
