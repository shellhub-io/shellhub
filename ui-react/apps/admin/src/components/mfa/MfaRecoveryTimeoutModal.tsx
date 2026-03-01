import { useState, useEffect } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useCountdown } from "../../hooks/useCountdown";

interface MfaRecoveryTimeoutModalProps {
  open: boolean;
  expiresAt: number; // Unix timestamp
  onClose: () => void;
  onDisable: () => Promise<void>;
}

export default function MfaRecoveryTimeoutModal({
  open,
  expiresAt,
  onClose,
  onDisable,
}: MfaRecoveryTimeoutModalProps) {
  const [hasAccess, setHasAccess] = useState(false);
  const [disabling, setDisabling] = useState(false);
  const { timeLeft, isExpired } = useCountdown(expiresAt);

  // Handle Escape key for accessibility
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !disabling) {
        onClose();
      }
    };

    if (open) {
      document.addEventListener("keydown", handleEscape);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
    };
  }, [open, onClose, disabling]);

  if (!open) return null;

  const handleDisable = async () => {
    setDisabling(true);
    try {
      await onDisable();
    } catch (error) {
      // Log error for debugging - parent component (MfaRecover) handles user feedback
      console.error("Failed to disable MFA during recovery window:", error);
      // Don't rethrow - allow modal to remain open for user to see parent's error or retry
    } finally {
      setDisabling(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      {/* Non-dismissible backdrop */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      <div
        className="relative bg-surface border border-border rounded-2xl w-full max-w-md mx-4 p-6 shadow-2xl animate-slide-up"
        role="dialog"
        aria-modal="true"
        aria-labelledby="recovery-timeout-title"
      >
        {/* Header */}
        <div className="flex items-start gap-3 mb-4">
          <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-accent-yellow/15 border border-accent-yellow/25 flex items-center justify-center">
            <ExclamationTriangleIcon
              className="w-5 h-5 text-accent-yellow"
              strokeWidth={2}
            />
          </div>
          <div>
            <h2
              id="recovery-timeout-title"
              className="text-base font-semibold text-text-primary mb-1"
            >
              Recovery Window Active
            </h2>
            <p className="text-xs font-mono text-accent-yellow">
              {isExpired ? "Expired" : `${timeLeft} remaining`}
            </p>
          </div>
        </div>

        {/* Description */}
        <div className="text-sm text-text-muted mb-6 leading-relaxed">
          <p className="mb-3">
            You've successfully used a recovery code. For security reasons, you
            now have a <strong className="text-text-primary">10-minute window</strong> to
            disable MFA if you no longer have access to your authenticator
            device.
          </p>
          <p className="text-xs">
            After this window expires, you'll need to use another recovery code
            or contact support.
          </p>
        </div>

        {/* Checkbox */}
        <label className="flex items-start gap-3 mb-6 cursor-pointer group">
          <input
            type="checkbox"
            checked={hasAccess}
            onChange={(e) => setHasAccess(e.target.checked)}
            className="mt-0.5 w-4 h-4 rounded border-border bg-background text-primary focus:ring-2 focus:ring-primary/20 cursor-pointer"
          />
          <span className="text-sm text-text-muted group-hover:text-text-secondary transition-colors">
            I have access to my authentication device and want to keep MFA
            enabled
          </span>
        </label>

        {/* Actions */}
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Close
          </button>
          <button
            onClick={handleDisable}
            disabled={hasAccess || disabling || isExpired}
            className="px-5 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {disabling && (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            )}
            Disable MFA
          </button>
        </div>

        {/* Explanation note */}
        <div className="mt-4 pt-4 border-t border-border">
          <p className="text-2xs text-text-muted leading-relaxed">
            <strong className="text-text-secondary">Why this window?</strong>{" "}
            This security measure prevents unauthorized access while allowing
            legitimate users to regain control if they've lost their
            authenticator device.
          </p>
        </div>
      </div>
    </div>
  );
}
