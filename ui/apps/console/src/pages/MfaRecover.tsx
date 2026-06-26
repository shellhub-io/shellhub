import { useState, FormEvent, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { KeyIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useAuthStore } from "../stores/authStore";
import { disableMfa } from "../client";
import MfaRecoveryTimeoutModal from "../components/mfa/MfaRecoveryTimeoutModal";
import AuthFooterLinks from "../components/common/AuthFooterLinks";

export default function MfaRecover() {
  const {
    recoverWithCode,
    loading,
    error,
    mfaRecoveryExpiry,
    updateMfaStatus,
    user,
    username,
    mfaToken,
  } = useAuthStore();
  const navigate = useNavigate();

  const identifier = user || username;
  const [recoveryCode, setRecoveryCode] = useState("");
  const [showTimeoutModal, setShowTimeoutModal] = useState(false);

  // Clear stale error from previous session
  useEffect(() => {
    useAuthStore.setState({ error: null });
  }, []);

  // Redirect to login if no identifier available (but only if not in active MFA session)
  useEffect(() => {
    if (!identifier && !mfaToken) {
      void navigate("/login");
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
    // Use the recovery code that was just entered
    await disableMfa({
      body: { recovery_code: recoveryCode },
      throwOnError: true,
    });
    updateMfaStatus(false);
    setShowTimeoutModal(false);
    void navigate("/dashboard");
  };

  const handleCloseModal = () => {
    setShowTimeoutModal(false);
    useAuthStore.setState({ mfaRecoveryExpiry: null });
    void navigate("/dashboard");
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
          Enter one of your recovery codes for{" "}
          <span className="font-semibold text-text-primary">{identifier}</span>.
        </p>
      </div>

      {/* Form card */}
      <div
        className="w-full max-w-sm bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
        style={{ animationDelay: "200ms" }}
      >
        <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
          {error && <Callout variant="error">{error}</Callout>}

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
              className="w-full px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-yellow/50 focus:ring-1 focus:ring-accent-yellow/20 transition-all duration-200"
              placeholder="Enter recovery code"
            />
            <p className="text-2xs text-text-muted mt-2">
              You received 6 recovery codes when you enabled MFA.
            </p>
          </div>

          <Button
            variant="warning"
            fullWidth
            type="submit"
            loading={loading}
            disabled={loading || !recoveryCode.trim()}
          >
            {loading ? "Recovering..." : "Recover Account"}
          </Button>

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
