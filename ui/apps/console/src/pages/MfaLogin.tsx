import { FormEvent, useEffect } from "react";
import { useNavigate, Link, useLocation } from "react-router-dom";
import { ShieldCheckIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";
import { useOtpInput } from "../hooks/useOtpInput";
import { resolvePostLoginRedirect } from "@/utils/navigation";
import PendingDeviceCallout from "@/components/auth/PendingDeviceCallout";
import { Button, Callout } from "@shellhub/design-system/primitives";
import AuthFooterLinks from "../components/common/AuthFooterLinks";
import LoginLayoutCard from "@/components/layout/LoginLayoutCard";

/**
 * The second factor at sign-in. Holds the partial token from the first step and exchanges it for
 * a full one, so a wrong code leaves the user still un-signed-in rather than signed out.
 */
export default function MfaLogin() {
  const otp = useOtpInput(6);
  const { loginWithMfa, loading, error, mfaToken } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!mfaToken) {
      void navigate("/login");
    }
  }, [mfaToken, navigate]);

  if (!mfaToken) {
    return null;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!otp.isComplete) return;

    try {
      await loginWithMfa(otp.getValue());
      const params = new URLSearchParams(location.search);
      void navigate(resolvePostLoginRedirect(params));
    } catch {
      otp.reset();
    }
  };

  return (
    <>
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

      <div className="w-full max-w-md flex flex-col gap-3 mb-4 empty:hidden">
        <PendingDeviceCallout />
      </div>

      {/* Form card */}
      <LoginLayoutCard>
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {error && <Callout variant="error">{error}</Callout>}

          <div>
            <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
              Verification Code
            </p>
            <div
              className="flex gap-2 justify-center"
              role="group"
              aria-label="Verification Code"
              onPaste={otp.handlePaste}
            >
              {otp.code.map((digit, index) => (
                <input
                  key={index}
                  ref={(el) => {
                    otp.inputRefs.current[index] = el;
                  }}
                  type="text"
                  inputMode="numeric"
                  maxLength={1}
                  value={digit}
                  aria-label={`Digit ${index + 1} of 6`}
                  onChange={(e) => otp.handleChange(index, e.target.value)}
                  onKeyDown={(e) => otp.handleKeyDown(index, e)}
                  className="w-12 h-12 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
                />
              ))}
            </div>
          </div>

          <Button
            variant="primary"
            size="lg"
            fullWidth
            type="submit"
            className="px-4"
            loading={loading}
            disabled={loading || !otp.isComplete}
          >
            {loading ? "Verifying..." : "Verify"}
          </Button>

          <div className="text-center pt-2">
            <Link
              to="/mfa-recover"
              className="text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              Lost your TOTP password?
            </Link>
          </div>
        </form>
      </LoginLayoutCard>
      {/* Footer links */}
      <AuthFooterLinks />
    </>
  );
}
