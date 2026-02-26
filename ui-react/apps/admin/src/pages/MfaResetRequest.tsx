import { FormEvent, useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { EnvelopeIcon, ExclamationCircleIcon, CheckCircleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaResetRequest() {
  const { requestMfaReset, loading, error, user, username, mfaToken } = useAuthStore();
  const navigate = useNavigate();
  const [emailsSent, setEmailsSent] = useState(false);

  const identifier = user || username;

  // Redirect to login if no identifier available (but only if not in active MFA session)
  useEffect(() => {
    if (!identifier && !mfaToken) {
      navigate("/login");
    }
  }, [identifier, mfaToken, navigate]);

  // Don't render if we don't have an identifier
  if (!identifier) {
    return null;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    try {
      await requestMfaReset(identifier);
      setEmailsSent(true);
    } catch {
      // Error is set in store
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero Section */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <EnvelopeIcon className="w-10 h-10 text-primary" strokeWidth={1.2} />
          </div>
        </div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Email Recovery
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Reset MFA via Email
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          We'll send verification codes to both email addresses registered for <span className="font-semibold text-text-primary">{identifier}</span>.
        </p>
      </div>

      {/* Form or Success Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        {!emailsSent ? (
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
                <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
                {error}
              </div>
            )}

            <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
              <p className="text-xs text-text-muted text-center">
                Verification codes will be sent to the email addresses registered for:
              </p>
              <p className="text-sm font-mono font-semibold text-primary text-center mt-2">
                {identifier}
              </p>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span className="font-mono text-xs">Sending...</span>
                </span>
              ) : (
                "Send Verification Codes"
              )}
            </button>

            <div className="text-center pt-2">
              <Link
                to="/mfa-recover"
                className="block text-xs text-text-muted hover:text-text-secondary transition-colors"
              >
                ← Back to recovery
              </Link>
            </div>
          </form>
        ) : (
          <div className="space-y-5 text-center">
            <div className="flex justify-center">
              <div className="w-16 h-16 rounded-full bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
                <CheckCircleIcon className="w-8 h-8 text-accent-green" strokeWidth={2} />
              </div>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-2">
                Emails Sent!
              </h3>
              <p className="text-sm text-text-muted leading-relaxed">
                We've sent verification codes to both your <span className="font-semibold text-text-primary">main email</span> and <span className="font-semibold text-text-primary">recovery email</span> addresses.
              </p>
            </div>

            <div className="p-4 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg">
              <p className="text-xs text-text-muted leading-relaxed">
                <span className="font-semibold text-accent-yellow">Next step:</span> Check both email inboxes and click the link in either email to continue with the reset process.
              </p>
            </div>

            <Link
              to="/login"
              className="block text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              ← Back to login
            </Link>
          </div>
        )}
      </div>

      {/* Info Note */}
      <div
        className="w-full max-w-sm mt-6 p-4 bg-primary/5 border border-primary/20 rounded-lg animate-fade-in"
        style={{ animationDelay: "400ms" }}
      >
        <p className="text-2xs text-text-muted leading-relaxed">
          <span className="font-semibold text-primary">Note:</span> You'll receive
          two separate emails with verification codes. Both codes are required to
          complete the reset process.
        </p>
      </div>

      <AuthFooterLinks />
    </div>
  );
}
