import { FormEvent, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  ExclamationCircleIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";
import { useOtpInput } from "../hooks/useOtpInput";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaLogin() {
  const otp = useOtpInput(6);
  const { loginWithMfa, loading, error, mfaToken } = useAuthStore();
  const navigate = useNavigate();

  // Redirect if no MFA token
  useEffect(() => {
    if (!mfaToken) {
      navigate("/login");
    }
  }, [mfaToken, navigate]);

  // Prevent rendering while redirecting
  if (!mfaToken) {
    return null;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otp.isComplete) return;

    try {
      await loginWithMfa(otp.getValue());
      navigate("/dashboard");
    } catch {
      // Error is set in store
      otp.reset();
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <ShieldCheckIcon
              className="w-10 h-10 text-primary"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Security Verification
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Two-Factor Authentication
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Enter the 6-digit code from your authenticator app to complete sign
          in.
        </p>
      </div>

      {/* Form card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <ExclamationCircleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          <div>
            <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
              Verification Code
            </label>
            <div className="flex gap-2 justify-center" onPaste={otp.handlePaste}>
              {otp.code.map((digit, index) => (
                <input
                  key={index}
                  ref={(el) => (otp.inputRefs.current[index] = el)}
                  type="text"
                  inputMode="numeric"
                  maxLength={1}
                  value={digit}
                  onChange={(e) => otp.handleChange(index, e.target.value)}
                  onKeyDown={(e) => otp.handleKeyDown(index, e)}
                  autoFocus={index === 0}
                  aria-label={`Digit ${index + 1} of 6`}
                  className="w-12 h-12 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
                />
              ))}
            </div>
          </div>

          <button
            type="submit"
            disabled={loading || !otp.isComplete}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span className="font-mono text-xs">Verifying...</span>
              </span>
            ) : (
              "Verify"
            )}
          </button>

          <div className="text-center pt-2">
            <Link
              to="/mfa-recover"
              className="text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              Lost your TOTP password?
            </Link>
          </div>
        </form>
      </div>

      {/* Footer links */}
      <AuthFooterLinks />
    </div>
  );
}
