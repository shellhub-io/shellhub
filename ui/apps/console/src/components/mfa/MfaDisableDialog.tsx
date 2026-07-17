import { useState, FormEvent } from "react";
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { disableMfa } from "@/client";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useOtpInput } from "@/hooks/useOtpInput";
import { useAuthStore } from "@/stores/authStore";
import { useMfaResetStore } from "@/stores/mfaResetStore";
import BaseDialog from "@/components/common/BaseDialog";

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
  const { user, username } = useAuthStore();
  const { requestMfaReset } = useMfaResetStore();

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
    setError("");
    setSubmitting(true);

    try {
      if (mode === "totp") {
        if (!otp.isComplete) return;
        await disableMfa({
          body: { code: otp.getValue() },
          throwOnError: true,
        });
      } else if (mode === "recovery") {
        if (!recoveryCode.trim()) return;
        await disableMfa({
          body: { recovery_code: recoveryCode },
          throwOnError: true,
        });
      } else if (mode === "email-reset") {
        if (!otpMainEmail.isComplete || !otpRecoveryEmail.isComplete) return;
        await disableMfa({
          body: {
            main_email_code: otpMainEmail.getValue(),
            recovery_email_code: otpRecoveryEmail.getValue(),
          },
          throwOnError: true,
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
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-label="Disable MFA"
    >
      <div className="p-6">
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

        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
          {error && <Callout variant="error">{error}</Callout>}

          {mode === "totp" ? (
            <>
              <div>
                <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
                  Verification Code
                </p>
                <div
                  className="flex gap-2 justify-center"
                  role="group"
                  aria-label="Verification Code"
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
                      onChange={(e) => otp.handleChange(index, e.target.value)}
                      onKeyDown={(e) => otp.handleKeyDown(index, e)}
                      aria-label={`Digit ${index + 1}`}
                      className="w-10 h-10 text-center text-base font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all"
                    />
                  ))}
                </div>
              </div>

              <div className="text-center">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setMode("recovery")}
                >
                  Use recovery code instead
                </Button>
              </div>
            </>
          ) : mode === "recovery" ? (
            <>
              <div>
                <label
                  htmlFor="disable-mfa-recovery-code"
                  className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2"
                >
                  Recovery Code
                </label>
                <input
                  id="disable-mfa-recovery-code"
                  type="text"
                  value={recoveryCode}
                  onChange={(e) => setRecoveryCode(e.target.value)}
                  className="w-full px-4 py-2.5 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all"
                  placeholder="Enter recovery code"
                />
              </div>

              <div className="text-center space-y-2">
                <Button
                  variant="ghost"
                  size="sm"
                  fullWidth
                  onClick={() => setMode("totp")}
                >
                  ← Use authenticator code
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  fullWidth
                  onClick={() => {
                    setMode("email-reset");
                    setEmailRequested(false);
                  }}
                >
                  Lost recovery codes? Request email reset
                </Button>
              </div>
            </>
          ) : (
            <>
              {!emailRequested ? (
                <div className="space-y-4">
                  <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
                    <p className="text-xs text-text-muted text-center">
                      Verification codes will be sent to both email addresses
                      registered for your account.
                    </p>
                  </div>

                  <Button
                    fullWidth
                    loading={requestingEmail}
                    onClick={() => void handleRequestEmailReset()}
                  >
                    Send Verification Codes
                  </Button>

                  <div className="text-center">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setMode("recovery")}
                    >
                      ← Use recovery code
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="flex items-center gap-2 justify-center">
                    <CheckCircleIcon
                      className="w-5 h-5 text-accent-green"
                      strokeWidth={2}
                    />
                    <p className="text-xs font-semibold text-text-primary">
                      Emails Sent!
                    </p>
                  </div>

                  <div>
                    <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5 text-center">
                      Main Email Code
                    </p>
                    <div
                      className="flex justify-center gap-2 mb-2"
                      role="group"
                      aria-label="Main Email Code"
                      onPaste={otpMainEmail.handlePaste}
                    >
                      {otpMainEmail.code.map((char, index) => (
                        <input
                          key={index}
                          ref={(el) => {
                            otpMainEmail.inputRefs.current[index] = el;
                          }}
                          type="text"
                          maxLength={1}
                          value={char}
                          aria-label={`Main email code character ${index + 1} of 5`}
                          onChange={(e) =>
                            otpMainEmail.handleChange(index, e.target.value)
                          }
                          onKeyDown={(e) =>
                            otpMainEmail.handleKeyDown(index, e)
                          }
                          className="w-10 h-10 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all uppercase"
                        />
                      ))}
                    </div>
                  </div>

                  <div>
                    <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5 text-center">
                      Recovery Email Code
                    </p>
                    <div
                      className="flex justify-center gap-2 mb-2"
                      role="group"
                      aria-label="Recovery Email Code"
                      onPaste={otpRecoveryEmail.handlePaste}
                    >
                      {otpRecoveryEmail.code.map((char, index) => (
                        <input
                          key={index}
                          ref={(el) => {
                            otpRecoveryEmail.inputRefs.current[index] = el;
                          }}
                          type="text"
                          maxLength={1}
                          value={char}
                          aria-label={`Recovery email code character ${index + 1} of 5`}
                          onChange={(e) =>
                            otpRecoveryEmail.handleChange(index, e.target.value)
                          }
                          onKeyDown={(e) =>
                            otpRecoveryEmail.handleKeyDown(index, e)
                          }
                          className="w-10 h-10 text-center text-lg font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-red/50 focus:ring-1 focus:ring-accent-red/20 transition-all uppercase"
                        />
                      ))}
                    </div>
                  </div>

                  <div className="text-center">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => void handleRequestEmailReset()}
                      loading={requestingEmail}
                    >
                      Resend codes
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}

          {/* Actions */}
          {(mode !== "email-reset" || emailRequested) && (
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="secondary" onClick={onClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                variant="destructive"
                disabled={!isComplete}
                loading={submitting}
              >
                Disable MFA
              </Button>
            </div>
          )}
        </form>
      </div>
    </BaseDialog>
  );
}
