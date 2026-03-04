import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  EnvelopeIcon,
  ArrowRightIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../../stores/authStore";
import { useSignUpStore } from "../../stores/signUpStore";
import { useResendEmail } from "../../hooks/useResendEmail";

interface AccountCreatedProps {
  mode: "normal" | "sig";
  username: string;
}

export default function AccountCreated({ mode, username }: AccountCreatedProps) {
  const navigate = useNavigate();
  const signUpToken = useSignUpStore((s) => s.signUpToken);
  const signUpTenant = useSignUpStore((s) => s.signUpTenant);
  const setSession = useAuthStore((s) => s.setSession);
  const { handleResend, resendLoading, resendError, resendSuccess, resendCooldown } = useResendEmail(username);

  useEffect(() => {
    if (mode !== "sig" || !signUpToken || !signUpTenant) return;

    setSession({ token: signUpToken, tenant: signUpTenant });

    // TODO: replace "/accept-invite" with the real route once it is added to App.tsx.
    const timer = setTimeout(() => { navigate("/accept-invite"); }, 5000);
    return () => clearTimeout(timer);
  }, [mode, signUpToken, signUpTenant, setSession, navigate]);

  const handleRedirect = () => {
    navigate("/accept-invite"); // TODO: update when accept-invite route is added
  };

  if (mode === "sig") {
    return (
      <div className="w-full max-w-sm mx-auto animate-fade-in">
        <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-accent-green/10 border border-accent-green/20 mb-5">
            <CheckCircleIcon className="w-7 h-7 text-accent-green" strokeWidth={1.5} />
          </div>

          <h2 className="text-lg font-semibold text-text-primary mb-3">
            Account Creation Successful
          </h2>

          <p className="text-sm text-text-secondary leading-relaxed mb-6">
            Thank you for registering an account on ShellHub. You will be
            redirected in 5 seconds. If you weren&apos;t redirected, please
            click the button below.
          </p>

          <button
            type="button"
            onClick={handleRedirect}
            className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200"
          >
            Redirect
            <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-sm mx-auto animate-fade-in">
      <div className="bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm text-center">
        <div className="inline-flex items-center justify-center w-14 h-14 rounded-full bg-primary/10 border border-primary/20 mb-5">
          <EnvelopeIcon className="w-7 h-7 text-primary" strokeWidth={1.5} />
        </div>

        <h2 className="text-lg font-semibold text-text-primary mb-3">
          Account Creation Successful
        </h2>

        <p className="text-sm text-text-secondary leading-relaxed mb-2">
          Thank you for registering an account on ShellHub. An email was sent
          with a confirmation link. You need to click on the link to activate
          your account.
        </p>

        <p className="text-xs text-text-muted mb-6">
          If you haven&apos;t received the email, click on the button.
        </p>

        {resendSuccess && (
          <div className="flex items-center gap-2 bg-accent-green/8 border border-accent-green/20 text-accent-green px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
            <CheckCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            Confirmation email sent successfully.
          </div>
        )}

        {resendError && (
          <div role="alert" className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            {resendError}
          </div>
        )}

        <button
          type="button"
          onClick={handleResend}
          disabled={resendLoading || resendCooldown > 0}
          className="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-white px-5 py-2.5 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200"
        >
          {resendLoading ? (
            <>
              <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              <span className="font-mono text-xs">Sending...</span>
            </>
          ) : resendCooldown > 0 ? (
            <span className="font-mono text-xs">Resend Email ({resendCooldown}s)</span>
          ) : (
            <>
              <EnvelopeIcon className="w-4 h-4" strokeWidth={2} />
              Resend Email
            </>
          )}
        </button>
      </div>
    </div>
  );
}
