import { useState, useEffect } from "react";
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
  const [loading, setLoading] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { handleDownload, handleCopy } = useRecoveryCodeActions();

  const loadCodes = async (): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      const data = await generateMfa();
      setCodes(data.recovery_codes);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load recovery codes";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && codes.length === 0) {
      void loadCodes();
    }
  }, [open, codes.length]);

  const handleRegenerate = async (): Promise<void> => {
    setRegenerating(true);
    setError(null);
    try {
      const data = await generateMfa();
      setCodes(data.recovery_codes);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to regenerate recovery codes";
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
              Use these codes to access your account if you lose your
              authenticator
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-4 flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3 py-2 rounded-md text-xs">
            <ExclamationTriangleIcon className="w-4 h-4 shrink-0" strokeWidth={2} />
            <span>{error}</span>
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <span className="w-6 h-6 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          </div>
        ) : (
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
                  Each recovery code can only be used once. Regenerating codes
                  will invalidate all previous codes.
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-between items-center gap-2 pt-2">
              <button
                onClick={handleRegenerate}
                disabled={regenerating}
                className="px-4 py-2 text-sm font-medium text-accent-yellow hover:text-accent-yellow/80 border border-accent-yellow/20 rounded-lg hover:bg-accent-yellow/5 transition-colors disabled:opacity-dim disabled:cursor-not-allowed flex items-center gap-2"
              >
                {regenerating && (
                  <span className="w-3.5 h-3.5 border-2 border-accent-yellow/30 border-t-accent-yellow rounded-full animate-spin" />
                )}
                Regenerate Codes
              </button>
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
