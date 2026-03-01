import { useState, FormEvent } from "react";
import { ExclamationTriangleIcon, CheckCircleIcon } from "@heroicons/react/24/outline";
import { disableMfa } from "../../api/mfa";
import { useOtpInput } from "../../hooks/useOtpInput";
import { useAuthStore } from "../../stores/authStore";

interface MfaDisableDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

type Mode = "totp" | "recovery" | "email-reset";

export default function MfaDisableDialog({
  open,
  onClose,
  onSuccess,
}: MfaDisableDialogProps) {
  const [mode, setMode] = useState<Mode>("totp");
  const otp = useOtpInput(6);
  const [recoveryCode, setRecoveryCode] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [emailRequested, setEmailRequested] = useState(false);
  const [requestingEmail, setRequestingEmail] = useState(false);
  const otpMainEmail = useOtpInput(5, true);
  const otpRecoveryEmail = useOtpInput(5, true);
  const { user, username, requestMfaReset } = useAuthStore();

  if (!open) return null;

  const handleRequestEmailReset = async (): Promise<void> => {
    const identifier = user || username;
    if (!identifier) {
      setError("Unable to identify user. Please try again.");
      return;
    }

    setRequestingEmail(true);
    setError("");
    try {
      await requestMfaReset(identifier);
      setEmailRequested(true);
    } catch {
      setError("Failed to send verification codes. Please try again.");
    } finally {
      setRequestingEmail(false);
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    // Validate input before setting loading state to avoid UI flicker
    if (mode === "totp" && !otp.isComplete) return;
    if (mode === "recovery" && !recoveryCode.trim()) return;
    if (mode === "email-reset" && (!otpMainEmail.isComplete || !otpRecoveryEmail.isComplete)) return;

    setError("");
    setSubmitting(true);

    try {
      if (mode === "totp") {
        await disableMfa({ code: otp.getValue() });
      } else if (mode === "recovery") {
        await disableMfa({ recovery_code: recoveryCode });
      } else if (mode === "email-reset") {
        await disableMfa({
          main_email_code: otpMainEmail.getValue(),
          recovery_email_code: otpRecoveryEmail.getValue(),
        });
      }

      onSuccess();
      onClose();
      // Reset state
      setTimeout(() => {
        setMode("totp");
        otp.reset();
        setRecoveryCode("");
        setEmailRequested(false);
        otpMainEmail.reset();
        otpRecoveryEmail.reset();
        setError("");
      }, 300);
    } catch {
      let errorMessage = "Invalid verification code";
      if (mode === "recovery") {
        errorMessage = "Invalid recovery code";
      } else if (mode === "email-reset") {
        errorMessage = "Invalid email verification codes";
      }
      setError(errorMessage);

      if (mode === "totp") {
        otp.reset();
      } else if (mode === "email-reset") {
        otpMainEmail.reset();
        otpRecoveryEmail.reset();
      }
    } finally {
      setSubmitting(false);
    }
  };

  const isComplete =
    mode === "totp"
      ? otp.isComplete
      : mode === "recovery"
      ? recoveryCode.trim() !== ""
      : otpMainEmail.isComplete && otpRecoveryEmail.isComplete;

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
        {/* Header */}
        <div className="flex items-start gap-3 mb-4">
          <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-accent-red/15 border border-accent-red/25 flex items-center justify-center">
            <ExclamationTriangleIcon
              className="w-5 h-5 text-accent-red"
              strokeWidth={2}
            />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              Disable MFA
            </h2>
            <p className="text-xs text-text-muted mt-0.5">
              This will reduce your account security
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono">
              <ExclamationTriangleIcon
                className="w-3.5 h-3.5 shrink-0"
                strokeWidth={2}
              />
              {error}
            </div>
          )}

          {mode === "totp" ? (
            <>
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
                      className="w-10 h-10 text-center text-base font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all"
                    />
                  ))}
                </div>
              </div>

              <div className="text-center">
                <button
                  type="button"
                  onClick={() => setMode("recovery")}
                  className="text-xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Use recovery code instead
                </button>
              </div>
            </>
          ) : mode === "recovery" ? (
            <>
              <div>
                <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
                  Recovery Code
                </label>
                <input
                  type="text"
                  value={recoveryCode}
                  onChange={(e) => setRecoveryCode(e.target.value)}
                  autoFocus
                  className="w-full px-4 py-2.5 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all"
                  placeholder="Enter recovery code"
                />
              </div>

              <div className="text-center space-y-2">
                <button
                  type="button"
                  onClick={() => setMode("totp")}
                  className="block w-full text-xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  ← Use authenticator code
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setMode("email-reset");
                    setEmailRequested(false);
                  }}
                  className="block w-full text-xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Lost recovery codes? Request email reset
                </button>
              </div>
            </>
          ) : (
            <>
              {!emailRequested ? (
                <div className="space-y-4">
                  <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
                    <p className="text-xs text-text-muted text-center">
                      Verification codes will be sent to both email addresses registered for your account.
                    </p>
                  </div>

                  <button
                    type="button"
                    onClick={handleRequestEmailReset}
                    disabled={requestingEmail}
                    className="w-full px-4 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all disabled:opacity-dim disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  >
                    {requestingEmail && (
                      <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    )}
                    Send Verification Codes
                  </button>

                  <div className="text-center">
                    <button
                      type="button"
                      onClick={() => setMode("recovery")}
                      className="text-xs text-text-muted hover:text-text-secondary transition-colors"
                    >
                      ← Use recovery code
                    </button>
                  </div>
                </div>
              ) : (
                <div className="space-y-4 text-center">
                  <div className="flex justify-center">
                    <div className="w-14 h-14 rounded-full bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
                      <CheckCircleIcon className="w-7 h-7 text-accent-green" strokeWidth={2} />
                    </div>
                  </div>

                  <div>
                    <h4 className="text-sm font-semibold text-text-primary mb-2">
                      Emails Sent!
                    </h4>
                    <p className="text-xs text-text-muted leading-relaxed">
                      Verification codes have been sent to both your main and recovery email addresses.
                    </p>
                  </div>

                  <div className="p-3 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg">
                    <p className="text-2xs text-text-muted leading-relaxed">
                      <span className="font-semibold text-accent-yellow">Next step:</span> Check both email inboxes and click the link in either email to continue.
                    </p>
                  </div>

                  <button
                    type="button"
                    onClick={onClose}
                    className="w-full px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
                  >
                    Close
                  </button>
                </div>
              )}
            </>
          )}

          {/* Actions */}
          {(mode !== "email-reset" || emailRequested) && (
            <div className="flex justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={submitting || !isComplete}
                className="px-5 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
              >
                {submitting && (
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                )}
                Disable MFA
              </button>
            </div>
          )}
        </form>
      </div>
    </div>
  );
}
