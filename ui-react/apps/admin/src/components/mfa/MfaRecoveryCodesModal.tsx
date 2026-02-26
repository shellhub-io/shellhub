import { useState } from "react";
import {
  KeyIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { generateMfa } from "../../api/mfa";
import { useRecoveryCodeActions } from "../../hooks/useRecoveryCodeActions";

interface MfaRecoveryCodesModalProps {
  open: boolean;
  onClose: () => void;
}

export default function MfaRecoveryCodesModal({
  open,
  onClose,
}: MfaRecoveryCodesModalProps) {
  const [codes, setCodes] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { handleDownload, handleCopy } = useRecoveryCodeActions();

  const handleRegenerateClick = () => {
    setShowConfirmation(true);
  };

  const handleConfirmRegenerate = async (): Promise<void> => {
    setRegenerating(true);
    setError(null);
    setShowConfirmation(false);
    try {
      // Generate new codes
      const data = await generateMfa();
      setCodes(data.recovery_codes);

      // TODO: Persist the new codes to the database
      // This requires either:
      // 1. A new backend endpoint specifically for regenerating codes
      // 2. Calling enableMfa with a dummy verification code during an active session
      // For now, these codes are generated but not persisted to the database
      // A proper implementation would call an endpoint here to save them

      // TEMPORARY: Log warning about persistence
      console.warn(
        "Recovery codes generated but not persisted. Backend changes needed to save regenerated codes."
      );
    } catch (err) {
      const message =
        err instanceof Error
          ? err.message
          : "Failed to regenerate recovery codes";
      setError(message);
    } finally {
      setRegenerating(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-md mx-4 p-6 shadow-2xl animate-slide-up">
        {/* Header */}
        <div className="flex items-start gap-3 mb-4">
          <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-accent-yellow/15 border border-accent-yellow/25 flex items-center justify-center">
            <KeyIcon className="w-5 h-5 text-accent-yellow" strokeWidth={2} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              Recovery Codes
            </h2>
            <p className="text-xs text-text-muted mt-0.5">
              {codes.length === 0
                ? "Regenerate your recovery codes"
                : "Use these codes to access your account if you lose your authenticator"}
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-4 flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3 py-2 rounded-md text-xs">
            <ExclamationTriangleIcon className="w-4 h-4 shrink-0" strokeWidth={2} />
            <span>{error}</span>
          </div>
        )}

        {showConfirmation ? (
          /* Confirmation Dialog */
          <div className="space-y-4">
            <div className="bg-accent-red/5 border border-accent-red/20 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <ExclamationTriangleIcon className="w-5 h-5 text-accent-red shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-semibold text-text-primary mb-2">
                    Are you sure you want to regenerate your recovery codes?
                  </p>
                  <p className="text-xs text-text-muted leading-relaxed">
                    This will invalidate all your current recovery codes. You'll
                    need to save the new codes in a safe place. Any old recovery
                    codes you have saved will no longer work.
                  </p>
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowConfirmation(false)}
                className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary border border-border rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmRegenerate}
                disabled={regenerating}
                className="px-4 py-2 bg-accent-red hover:bg-accent-red/80 text-white rounded-lg text-sm font-semibold transition-all disabled:opacity-dim disabled:cursor-not-allowed flex items-center gap-2"
              >
                {regenerating && (
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                )}
                Regenerate Codes
              </button>
            </div>
          </div>
        ) : codes.length === 0 ? (
          /* Empty State */
          <div className="space-y-4">
            <div className="bg-background border border-border rounded-lg p-6 text-center">
              <KeyIcon className="w-12 h-12 text-text-muted mx-auto mb-3 opacity-50" />
              <p className="text-sm text-text-muted mb-1">
                Recovery codes cannot be viewed after creation
              </p>
              <p className="text-xs text-text-muted leading-relaxed">
                For security reasons, recovery codes are hashed and cannot be
                retrieved. You can regenerate new codes, which will invalidate any
                existing codes.
              </p>
            </div>

            <div className="flex justify-end gap-2">
              <button
                onClick={onClose}
                className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary border border-border rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Close
              </button>
              <button
                onClick={handleRegenerateClick}
                className="px-4 py-2 bg-accent-yellow hover:bg-accent-yellow/80 text-background rounded-lg text-sm font-semibold transition-all"
              >
                Regenerate Codes
              </button>
            </div>
          </div>
        ) : (
          /* Codes Display */
          <div className="space-y-4">
            {/* Codes Grid */}
            <div className="bg-background border border-border rounded-lg p-4">
              <div className="grid grid-cols-2 gap-2 mb-3">
                {codes.map((code, index) => (
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
                  onClick={() => handleDownload(codes)}
                  className="flex-1 px-3 py-1.5 text-xs font-medium text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-hover-subtle transition-colors"
                >
                  Download
                </button>
                <button
                  onClick={() => handleCopy(codes)}
                  className="flex-1 px-3 py-1.5 text-xs font-medium text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-hover-subtle transition-colors"
                >
                  Copy
                </button>
              </div>
            </div>

            {/* Warning */}
            <div className="bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg p-3">
              <div className="flex items-start gap-2">
                <ExclamationTriangleIcon className="w-4 h-4 text-accent-yellow shrink-0 mt-0.5" />
                <p className="text-2xs text-text-muted leading-relaxed">
                  <span className="font-semibold text-accent-yellow">
                    Important:
                  </span>{" "}
                  Save these codes now! Each code can only be used once. Store
                  them in a secure location like a password manager.
                </p>
              </div>
            </div>

            {/* Info about persistence */}
            <div className="bg-accent-red/5 border border-accent-red/20 rounded-lg p-3">
              <div className="flex items-start gap-2">
                <ExclamationTriangleIcon className="w-4 h-4 text-accent-red shrink-0 mt-0.5" />
                <p className="text-2xs text-text-muted leading-relaxed">
                  <span className="font-semibold text-accent-red">
                    Note:
                  </span>{" "}
                  These regenerated codes are displayed but may not be persisted to
                  the database yet. Contact your administrator if you encounter
                  issues using them.
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={onClose}
                className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
              >
                Done
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
