import { useState, FormEvent, useEffect, useRef } from "react";
import {
  ShieldCheckIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  EnvelopeIcon,
} from "@heroicons/react/24/outline";
import Drawer from "../common/Drawer";
import { QRCodeDisplay } from "./QRCodeDisplay";
import { generateMfa, enableMfa } from "../../api/mfa";
import { updateUser } from "../../api/auth";
import type { MfaGenerateResponse } from "../../types/mfa";
import { useOtpInput } from "../../hooks/useOtpInput";
import { useRecoveryCodeActions } from "../../hooks/useRecoveryCodeActions";

interface MfaEnableDrawerProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  currentRecoveryEmail: string | null;
}

type Step = 1 | 2 | 3 | 4;

export default function MfaEnableDrawer({
  open,
  onClose,
  onSuccess,
  currentRecoveryEmail,
}: MfaEnableDrawerProps) {
  const [step, setStep] = useState<Step>(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Step 1: Recovery Email
  const [recoveryEmail, setRecoveryEmail] = useState("");
  const [showRecoveryEmailInput, setShowRecoveryEmailInput] = useState(!currentRecoveryEmail);

  // Step 2: Recovery Codes
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [codesSaved, setCodesSaved] = useState(false);

  // Step 3: QR Code + Verification
  const [qrLink, setQrLink] = useState("");
  const [secret, setSecret] = useState("");
  const otp = useOtpInput(6);
  const { handleDownload, handleCopy } = useRecoveryCodeActions();

  // Sync showRecoveryEmailInput when drawer opens or currentRecoveryEmail changes
  useEffect(() => {
    if (open) {
      setShowRecoveryEmailInput(!currentRecoveryEmail);
    }
  }, [open, currentRecoveryEmail]);

  const handleConfirmExistingEmail = async () => {
    setError("");
    setLoading(true);
    try {
      await handleGenerateMfa();
      setStep(2);
    } catch {
      setError("Failed to generate MFA codes");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveRecoveryEmail = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await updateUser({ recovery_email: recoveryEmail });
      await handleGenerateMfa();
      setStep(2);
    } catch (err) {
      if (err instanceof Error && err.message.includes("409")) {
        setError("Email already in use");
      } else {
        setError("Failed to save recovery email");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateMfa = async () => {
    try {
      const data: MfaGenerateResponse = await generateMfa();
      setQrLink(data.link);
      setSecret(data.secret);
      setRecoveryCodes(data.recovery_codes);
    } catch (error) {
      setError("Failed to generate MFA codes");
      throw error; // Re-throw so callers know it failed
    }
  };

  const handleNextToQr = async () => {
    if (!codesSaved) return;
    setError("");

    // If we skipped step 1 (already had recovery email), generate codes now
    if (!qrLink) {
      setLoading(true);
      try {
        await handleGenerateMfa();
        setStep(3);
      } catch {
        setError("Failed to generate QR code");
      } finally {
        setLoading(false);
      }
    } else {
      setStep(3);
    }
  };



  const handleEnableMfa = async (e: FormEvent) => {
    e.preventDefault();
    if (!otp.isComplete) return;

    setError("");
    setLoading(true);

    try {
      await enableMfa({ code: otp.getValue(), secret, recovery_codes: recoveryCodes });
      setStep(4);
    } catch {
      setError("Invalid verification code");
      otp.reset();
    } finally {
      setLoading(false);
    }
  };

  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const handleDone = (): void => {
    onSuccess();
    onClose();
    // Reset state after drawer animation completes
    timeoutRef.current = setTimeout(() => {
      setStep(1);
      setRecoveryEmail("");
      setShowRecoveryEmailInput(!currentRecoveryEmail);
      setRecoveryCodes([]);
      setCodesSaved(false);
      setQrLink("");
      setSecret("");
      otp.reset();
      setError("");
    }, 300);
  };

  const isCodeComplete = otp.isComplete;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Enable MFA"
      subtitle="Secure your account with two-factor authentication"
      icon={<ShieldCheckIcon className="w-5 h-5 text-primary" />}
      width="md"
      footer={
        step === 1 ? (
          <>
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            {showRecoveryEmailInput ? (
              <button
                onClick={handleSaveRecoveryEmail}
                disabled={loading || !recoveryEmail.trim()}
                className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
              >
                {loading && (
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                )}
                Save & Continue
              </button>
            ) : (
              <button
                onClick={handleConfirmExistingEmail}
                disabled={loading}
                className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
              >
                {loading && (
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                )}
                Continue
              </button>
            )}
          </>
        ) : step === 2 ? (
          <>
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleNextToQr}
              disabled={!codesSaved || loading || recoveryCodes.length === 0}
              className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
            >
              Next Step
            </button>
          </>
        ) : step === 3 ? (
          <>
            <button
              onClick={() => setStep(2)}
              className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Back
            </button>
            <button
              onClick={handleEnableMfa}
              disabled={loading || !isCodeComplete || !qrLink || !secret}
              className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {loading && (
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              )}
              Verify & Enable
            </button>
          </>
        ) : (
          <button
            onClick={handleDone}
            className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            Done
          </button>
        )
      }
    >
      <div className="space-y-5">
        {error && (
          <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </div>
        )}

        {/* Step 1: Recovery Email */}
        {step === 1 && (
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-2">
                Step 1: {showRecoveryEmailInput ? 'Set Recovery Email' : 'Confirm Recovery Email'}
              </h3>
              <p className="text-xs text-text-muted leading-relaxed mb-4">
                {showRecoveryEmailInput
                  ? 'This email will be used to recover your account if you lose access to your authenticator device.'
                  : 'Continue enabling MFA with your current recovery email, or change it below.'}
              </p>
            </div>

            {showRecoveryEmailInput ? (
              <div>
                <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
                  Recovery Email
                </label>
                <input
                  type="email"
                  value={recoveryEmail}
                  onChange={(e) => setRecoveryEmail(e.target.value)}
                  required
                  autoFocus
                  className="w-full px-4 py-2.5 bg-background border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                  placeholder="recovery@example.com"
                />
                <p className="text-2xs text-text-muted mt-1.5">
                  Must be different from your main email address
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="bg-surface border border-border rounded-lg p-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
                      <EnvelopeIcon className="w-5 h-5 text-primary" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1">
                        Current Recovery Email
                      </p>
                      <p className="text-sm font-mono text-text-primary truncate">
                        {currentRecoveryEmail}
                      </p>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => setShowRecoveryEmailInput(true)}
                  className="w-full px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary border border-border rounded-lg hover:bg-hover-subtle transition-colors"
                >
                  Use a different recovery email
                </button>
              </div>
            )}
          </div>
        )}

        {/* Step 2: Save Recovery Codes */}
        {step === 2 && (
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-2">
                Step 2: Save Recovery Codes
              </h3>
              <p className="text-xs text-text-muted leading-relaxed mb-4">
                Save these codes in a secure location. You'll need them to
                recover access if you lose your authenticator device.
              </p>
            </div>

            <div className="bg-background border border-border rounded-lg p-4">
              {recoveryCodes.length > 0 ? (
                <>
                  <div className="grid grid-cols-2 gap-2 mb-3">
                    {recoveryCodes.map((code, index) => (
                      <div
                        key={index}
                        className="px-3 py-2 bg-surface border border-border rounded text-xs font-mono text-text-primary text-center"
                      >
                        {code}
                      </div>
                    ))}
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleDownload(recoveryCodes)}
                      className="flex-1 px-3 py-1.5 text-xs font-medium text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-hover-subtle transition-colors"
                    >
                      Download
                    </button>
                    <button
                      onClick={() => handleCopy(recoveryCodes)}
                      className="flex-1 px-3 py-1.5 text-xs font-medium text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-hover-subtle transition-colors"
                    >
                      Copy
                    </button>
                  </div>
                </>
              ) : (
                <div className="py-8 text-center">
                  <p className="text-sm text-text-muted">
                    Recovery codes failed to generate. Please close and try again.
                  </p>
                </div>
              )}
            </div>

            <div className="bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg p-3">
              <p className="text-2xs text-text-muted leading-relaxed">
                <span className="font-semibold text-accent-yellow">
                  Important:
                </span>{" "}
                Each recovery code can only be used once. Store them securely —
                you won't be able to see them again.
              </p>
            </div>

            <label className="flex items-start gap-3 cursor-pointer group">
              <input
                type="checkbox"
                checked={codesSaved}
                onChange={(e) => setCodesSaved(e.target.checked)}
                className="mt-0.5 w-4 h-4 rounded border-border bg-background text-primary focus:ring-2 focus:ring-primary/20 cursor-pointer"
              />
              <span className="text-sm text-text-muted group-hover:text-text-secondary transition-colors">
                I have saved my recovery codes in a secure location
              </span>
            </label>
          </div>
        )}

        {/* Step 3: QR Code + Verification */}
        {step === 3 && (
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-2">
                Step 3: Scan QR Code
              </h3>
              <p className="text-xs text-text-muted leading-relaxed mb-4">
                Scan this QR code with your authenticator app (Google
                Authenticator, Authy, etc.) and enter the 6-digit code to
                verify.
              </p>
            </div>

            <div className="bg-background border border-border rounded-lg p-4 space-y-4">
              {qrLink && secret ? (
                <>
                  <div className="flex justify-center">
                    <QRCodeDisplay data={qrLink} size={180} />
                  </div>

                  <div>
                    <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
                      Manual Entry Key
                    </label>
                    <input
                      type="text"
                      value={secret}
                      readOnly
                      className="w-full px-3 py-2 bg-surface border border-border rounded text-xs font-mono text-text-primary text-center select-all"
                    />
                    <p className="text-2xs text-text-muted mt-1.5">
                      Use this key if you can't scan the QR code
                    </p>
                  </div>
                </>
              ) : (
                <div className="py-8 text-center">
                  <p className="text-sm text-text-muted">
                    QR code failed to generate. Please go back and try again.
                  </p>
                </div>
              )}
            </div>

            <div>
              <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3 text-center">
                Verification Code
              </label>
              <div className="flex gap-2 justify-center">
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
                    className="w-10 h-10 text-center text-base font-mono bg-background border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                  />
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Step 4: Success */}
        {step === 4 && (
          <div className="text-center py-8 space-y-4">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-accent-green/15 border border-accent-green/25">
              <CheckCircleIcon className="w-8 h-8 text-accent-green" />
            </div>

            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-2">
                MFA Enabled Successfully!
              </h3>
              <p className="text-sm text-text-muted leading-relaxed max-w-sm mx-auto">
                Your account is now protected with two-factor authentication.
              </p>
            </div>

            <div className="bg-surface border border-border rounded-lg p-4 text-left space-y-2">
              <p className="text-xs font-semibold text-text-primary">
                What's next?
              </p>
              <ul className="space-y-1.5 text-xs text-text-muted">
                <li className="flex items-start gap-2">
                  <span className="text-accent-green mt-0.5">•</span>
                  <span>
                    You'll be prompted for a code from your authenticator app at
                    every login
                  </span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-accent-green mt-0.5">•</span>
                  <span>
                    Keep your recovery codes safe — they're your backup access
                  </span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-accent-green mt-0.5">•</span>
                  <span>
                    To regenerate recovery codes, disable MFA and re-enable it
                  </span>
                </li>
              </ul>
            </div>
          </div>
        )}
      </div>
    </Drawer>
  );
}
