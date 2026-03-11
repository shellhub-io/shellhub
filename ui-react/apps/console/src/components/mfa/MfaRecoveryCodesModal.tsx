import {
  KeyIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";

interface MfaRecoveryCodesModalProps {
  open: boolean;
  onClose: () => void;
}

export default function MfaRecoveryCodesModal({
  open,
  onClose,
}: MfaRecoveryCodesModalProps) {
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

        <div className="space-y-4">
          <div className="bg-accent-yellow/5 border border-accent-yellow/20 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <ExclamationTriangleIcon className="w-5 h-5 text-accent-yellow shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-semibold text-text-primary mb-1">
                  Recovery codes cannot be viewed
                </p>
                <p className="text-xs text-text-muted leading-relaxed">
                  For security, recovery codes are only shown once during MFA
                  setup. To get new recovery codes, disable MFA and re-enable
                  it.
                </p>
              </div>
            </div>
          </div>

          <div className="flex justify-end pt-2">
            <button
              onClick={onClose}
              className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
