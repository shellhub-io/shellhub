import { FormEvent, useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ShieldCheckIcon, ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { completeMfaReset } from "../api/mfa";
import { useAuthStore } from "../stores/authStore";
import { useOtpInput } from "../hooks/useOtpInput";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaResetComplete() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setSession, updateMfaStatus } = useAuthStore();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const otpMain = useOtpInput(5, true);
  const otpRecovery = useOtpInput(5, true);

  const userId = searchParams.get("id");

  // Redirect if no user ID in URL
  useEffect(() => {
    if (!userId) {
      navigate("/login");
    }
  }, [userId, navigate]);

  if (!userId) {
    return null;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otpMain.isComplete || !otpRecovery.isComplete) return;

    setLoading(true);
    setError(null);

    try {
      const data = await completeMfaReset(userId, {
        main_email_code: otpMain.getValue(),
        recovery_email_code: otpRecovery.getValue(),
      });

      // Successful reset = authenticated
      setSession({ token: data.token, tenant: data.tenant });
      updateMfaStatus(data.mfa || false);
      navigate("/dashboard");
    } catch {
      setError("Invalid verification codes. Please check and try again.");
      otpMain.reset();
      otpRecovery.reset();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero Section */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
            <ShieldCheckIcon className="w-10 h-10 text-primary" strokeWidth={1.2} />
          </div>
        </div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
          Email Verification
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Enter Verification Codes
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Enter the codes from both emails to complete the MFA reset.
        </p>
      </div>

      {/* Form Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
              <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
              {error}
            </div>
          )}

          {/* Main Email Code */}
          <div>
            <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
              Main Email Code
            </label>
            <div className="flex justify-center gap-2 mb-2" onPaste={otpMain.handlePaste}>
              {otpMain.code.map((char, index) => (
                <input
                  key={index}
                  ref={(el) => (otpMain.inputRefs.current[index] = el)}
                  type="text"
                  maxLength={1}
                  value={char}
                  onChange={(e) => otpMain.handleChange(index, e.target.value)}
                  onKeyDown={(e) => otpMain.handleKeyDown(index, e)}
                  autoFocus={index === 0}
                  className="w-10 h-10 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 uppercase"
                />
              ))}
            </div>
            <p className="text-2xs text-text-muted text-center">
              Code sent to your main email address
            </p>
          </div>

          {/* Recovery Email Code */}
          <div>
            <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
              Recovery Email Code
            </label>
            <div className="flex justify-center gap-2 mb-2" onPaste={otpRecovery.handlePaste}>
              {otpRecovery.code.map((char, index) => (
                <input
                  key={index}
                  ref={(el) => (otpRecovery.inputRefs.current[index] = el)}
                  type="text"
                  maxLength={1}
                  value={char}
                  onChange={(e) => otpRecovery.handleChange(index, e.target.value)}
                  onKeyDown={(e) => otpRecovery.handleKeyDown(index, e)}
                  className="w-10 h-10 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 uppercase"
                />
              ))}
            </div>
            <p className="text-2xs text-text-muted text-center">
              Code sent to your recovery email address
            </p>
          </div>

          <button
            type="submit"
            disabled={!otpMain.isComplete || !otpRecovery.isComplete || loading}
            className="w-full bg-primary hover:bg-primary-600 text-white py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span className="font-mono text-xs">Verifying...</span>
              </span>
            ) : (
              "Reset MFA and Login"
            )}
          </button>
        </form>
      </div>

      {/* Info Note */}
      <div
        className="w-full max-w-sm mt-6 p-4 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg animate-fade-in"
        style={{ animationDelay: "400ms" }}
      >
        <p className="text-2xs text-text-muted leading-relaxed">
          <span className="font-semibold text-accent-yellow">Security:</span> Both
          codes are required to prove ownership of your account's email addresses.
          Codes expire after 24 hours.
        </p>
      </div>

      <AuthFooterLinks />
    </div>
  );
}
