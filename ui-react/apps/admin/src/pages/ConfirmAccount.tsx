import { useEffect } from "react";
import { Link, Navigate, useSearchParams } from "react-router-dom";
import {
  EnvelopeIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useSignUpStore } from "../stores/signUpStore";
import { useResendEmail } from "../hooks/useResendEmail";

export default function ConfirmAccount() {
  const [searchParams] = useSearchParams();
  const resetErrors = useSignUpStore((s) => s.resetResendError);

  const username = searchParams.get("username") ?? "";

  useEffect(() => {
    resetErrors();
  }, [resetErrors]);

  const { handleResend, resendLoading, resendError, resendSuccess, resendCooldown } = useResendEmail(username);

  if (!username) return <Navigate to="/login" replace />;

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center animate-slide-up"
      >
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
          <div className="flex items-center gap-2 bg-accent-green/8 border border-accent-green/20 text-accent-green px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down mb-4">
            <CheckCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            Confirmation email sent successfully.
          </div>
        )}

        {resendError && (
          <div role="alert" className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down mb-4">
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            {resendError}
          </div>
        )}

        <button
          type="button"
          onClick={handleResend}
          disabled={resendLoading || resendCooldown > 0}
          className="w-full bg-primary hover:bg-primary/90 text-white py-2.5 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mb-5"
        >
          {resendLoading ? (
            <span className="flex items-center justify-center gap-2">
              <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              <span className="font-mono text-xs">Sending...</span>
            </span>
          ) : resendCooldown > 0 ? (
            `Resend Email (${resendCooldown}s)`
          ) : (
            "Resend Email"
          )}
        </button>

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
