import { useState, FormEvent, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  ExclamationCircleIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "../stores/authStore";
import { disableMfa } from "../api/mfa";
import MfaRecoveryTimeoutModal from "../components/mfa/MfaRecoveryTimeoutModal";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaRecover() {
  const { recoverWithCode, loading, error, mfaRecoveryExpiry, updateMfaStatus, user, username, mfaToken, pendingMfaUser } = useAuthStore();
  const navigate = useNavigate();

  const identifier = pendingMfaUser || user || username;
  const [recoveryCode, setRecoveryCode] = useState("");
  const [showTimeoutModal, setShowTimeoutModal] = useState(false);

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
    if (!recoveryCode.trim()) return;

    try {
      await recoverWithCode(recoveryCode, identifier);
      setShowTimeoutModal(true);
    } catch {
      // Error is set in store
      setRecoveryCode("");
    }
  };

  const handleDisableMfa = async () => {
    try {
      // During recovery window, backend validates via session token
      // No need to re-send the already-consumed recovery code
      await disableMfa({});
      updateMfaStatus(false);
      setShowTimeoutModal(false);
      navigate("/dashboard");
    } catch (err) {
      // Show error to user - they may need to re-authenticate
      console.error("Failed to disable MFA:", err);
      // Close modal and let user see error from the API
      setShowTimeoutModal(false);
    }
  };

  const handleCloseModal = () => {
    setShowTimeoutModal(false);
    navigate("/dashboard");
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col items-center">
      {/* Hero */}
      <div className="text-center mb-12 animate-fade-in">
        <div className="animate-float mb-6 inline-block">
          <div className="w-20 h-20 rounded-2xl bg-accent-yellow/15 border border-accent-yellow/25 flex items-center justify-center shadow-lg shadow-accent-yellow/10">
            <KeyIcon
              className="w-10 h-10 text-accent-yellow"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-accent-yellow/80 mb-2">
          Account Recovery
        </p>
        <h1 className="text-3xl font-bold text-text-primary mb-3">
          Recover Your Account
        </h1>
        <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
          Enter one of your recovery codes for <span className="font-semibold text-text-primary">{identifier}</span>.
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
            <label
              htmlFor="recovery-code"
              className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5"
            >
              Recovery Code
            </label>
            <input
              id="recovery-code"
              type="text"
              value={recoveryCode}
              onChange={(e) => setRecoveryCode(e.target.value)}
              required
              autoFocus
              autoComplete="off"
              spellCheck={false}
              className="w-full px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-yellow/50 focus:ring-1 focus:ring-accent-yellow/20 transition-all duration-200"
              placeholder="Enter recovery code"
            />
            <p className="text-2xs text-text-muted mt-2">
              You received 6 recovery codes when you enabled MFA.
            </p>
          </div>

          <button
            type="submit"
            disabled={loading || !recoveryCode.trim()}
            className="w-full bg-accent-yellow hover:bg-accent-yellow/80 text-background py-3 px-4 rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 mt-1"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-background/30 border-t-background rounded-full animate-spin" />
                <span className="font-mono text-xs">Recovering...</span>
              </span>
            ) : (
              "Recover Account"
            )}
          </button>

          <div className="text-center pt-2 space-y-2">
            <Link
              to="/mfa-login"
              className="block text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              ← Back to verification
            </Link>
            <Link
              to="/mfa-reset-request"
              className="block text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              Lost recovery codes? Request email reset
            </Link>
          </div>
        </form>
      </div>

      {/* Warning note */}
      <div
        className="w-full max-w-sm mt-6 p-4 bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg animate-fade-in"
        style={{ animationDelay: "400ms" }}
      >
        <p className="text-2xs text-text-muted leading-relaxed">
          <span className="font-semibold text-accent-yellow">Note:</span> After
          using a recovery code, you'll have a 10-minute window to disable MFA
          if you no longer have access to your authenticator device.
        </p>
      </div>

      {/* Footer links */}
      <AuthFooterLinks />

      {/* Timeout Modal */}
      {showTimeoutModal && mfaRecoveryExpiry && (
        <MfaRecoveryTimeoutModal
          open={showTimeoutModal}
          expiresAt={mfaRecoveryExpiry}
          onClose={handleCloseModal}
          onDisable={handleDisableMfa}
        />
      )}
    </div>
  );
}
