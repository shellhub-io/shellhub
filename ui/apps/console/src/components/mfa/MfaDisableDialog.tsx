import { useState, FormEvent } from "react";
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { disableMfa } from "@/client";
import Alert from "@/components/common/Alert";
import { useOtpInput } from "@/hooks/useOtpInput";
import { useAuthStore } from "@/stores/authStore";
import { useMfaResetStore } from "@/stores/mfaResetStore";
import { Button } from "@shellhub/design-system/primitives";
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
          {error && <Alert variant="error">{error}</Alert>}

          {mode === "totp" ? (
            <>
              <div>
                <p className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
                  Verification Code
                </p>
                <div className="flex gap-2 justify-center" role="group" aria-label="Verification Code">
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
                <div className="space-y-4 text-center">
                  <div className="flex justify-center">
                    <div className="w-14 h-14 rounded-full bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
                      <CheckCircleIcon
                        className="w-7 h-7 text-accent-green"
                        strokeWidth={2}
                      />
                    </div>
                  </div>

                  <div>
                    <h4 className="text-sm font-semibold text-text-primary mb-2">
                      Emails Sent!
                    </h4>
                    <p className="text-xs text-text-muted leading-relaxed">
                      Verification codes have been sent to both your main and
                      recovery email addresses.
                    </p>
                  </div>

                  <div className="p-3 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg">
                    <p className="text-2xs text-text-muted leading-relaxed">
                      <span className="font-semibold text-accent-yellow">
                        Next step:
                      </span>{" "}
                      Check both email inboxes and click the link in either
                      email to continue.
                    </p>
                  </div>

                  <Button fullWidth onClick={onClose}>
                    Close
                  </Button>
                </div>
              )}
            </>
          )}

          {/* Actions */}
          {mode !== "email-reset" && (
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
