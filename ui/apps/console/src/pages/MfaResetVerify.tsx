import { FormEvent, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { ShieldCheckIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useMfaResetStore } from "../stores/mfaResetStore";
import { useOtpInput } from "../hooks/useOtpInput";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaResetVerify() {
  const {
    completeMfaReset,
    mfaResetUserId,
    mfaResetIdentifier,
    loading,
    error,
  } = useMfaResetStore();
  const navigate = useNavigate();

  const otpMain = useOtpInput(5, true);
  const otpRecovery = useOtpInput(5, true);

  // Clear stale error from previous session
  useEffect(() => {
    useMfaResetStore.setState({ error: null });
  }, []);

  // State guard: redirect if no reset session
  useEffect(() => {
    if (!mfaResetUserId) {
      void navigate("/mfa-recover");
    }
  }, [mfaResetUserId, navigate]);

  if (!mfaResetUserId) {
    return null; // Prevent rendering while redirecting
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otpMain.isComplete || !otpRecovery.isComplete) return;

    try {
      await completeMfaReset(otpMain.getValue(), otpRecovery.getValue());
      void navigate("/dashboard");
    } catch {
      // Error is set in store
      otpMain.reset();
      otpRecovery.reset();
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero Section */}
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
          Email Verification
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Enter Verification Codes
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Check both email addresses for {mfaResetIdentifier} and enter the
          codes below.
        </p>
      </div>

      {/* Form Card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {error && <Callout variant="error">{error}</Callout>}

          {/* Main Email Code */}
          <div>
            <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
              Main Email Code
            </p>
            <div
              className="flex justify-center gap-2 mb-2"
              role="group"
              aria-label="Main Email Code"
              onPaste={otpMain.handlePaste}
            >
              {otpMain.code.map((char, index) => (
                <input
                  key={index}
                  ref={(el) => (otpMain.inputRefs.current[index] = el)}
                  type="text"
                  maxLength={1}
                  value={char}
                  aria-label={`Main email code character ${index + 1} of 5`}
                  onChange={(e) => otpMain.handleChange(index, e.target.value)}
                  onKeyDown={(e) => otpMain.handleKeyDown(index, e)}
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
            <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
              Recovery Email Code
            </p>
            <div
              className="flex justify-center gap-2 mb-2"
              role="group"
              aria-label="Recovery Email Code"
              onPaste={otpRecovery.handlePaste}
            >
              {otpRecovery.code.map((char, index) => (
                <input
                  key={index}
                  ref={(el) => (otpRecovery.inputRefs.current[index] = el)}
                  type="text"
                  maxLength={1}
                  value={char}
                  aria-label={`Recovery email code character ${index + 1} of 5`}
                  onChange={(e) =>
                    otpRecovery.handleChange(index, e.target.value)
                  }
                  onKeyDown={(e) => otpRecovery.handleKeyDown(index, e)}
                  className="w-10 h-10 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200 uppercase"
                />
              ))}
            </div>
            <p className="text-2xs text-text-muted text-center">
              Code sent to your recovery email address
            </p>
          </div>

          <Button
            variant="primary"
            size="lg"
            fullWidth
            type="submit"
            className="px-4"
            loading={loading}
            disabled={!otpMain.isComplete || !otpRecovery.isComplete || loading}
          >
            {loading ? "Verifying..." : "Verify and Reset MFA"}
          </Button>

          <div className="text-center pt-2">
            <Link
              to="/mfa-reset-request"
              className="block text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              ← Resend codes
            </Link>
          </div>
        </form>
      </div>

      {/* Info Note */}
      <div
        className="w-full max-w-sm mt-6 p-4 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg animate-fade-in"
        style={{ animationDelay: "400ms" }}
      >
        <p className="text-2xs text-text-muted leading-relaxed">
          <span className="font-semibold text-accent-yellow">Security:</span>{" "}
          Both codes are required to prove ownership of your account's email
          addresses. Codes expire after 24 hours.
        </p>
      </div>

      <AuthFooterLinks />
    </div>
  );
}
