import { useState } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useCountdown } from "@/hooks/useCountdown";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { Button } from "@shellhub/design-system/primitives";
import BaseDialog from "@/components/common/BaseDialog";

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

  const handleDisable = async () => {
    setDisabling(true);
    try {
      await onDisable();
    } finally {
      setDisabling(false);
    }
  };

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      canClose={() => false}
      size="md"
      aria-label="Recovery Window Active"
    >
      <div className="p-6">
        {/* Header */}
        <div className="flex items-start gap-3 mb-4">
          <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-accent-yellow/15 border border-accent-yellow/25 flex items-center justify-center">
            <ExclamationTriangleIcon
              className="w-5 h-5 text-accent-yellow"
              strokeWidth={2}
            />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary mb-1">
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
            now have a{" "}
            <strong className="text-text-primary">10-minute window</strong> to
            disable MFA if you no longer have access to your authenticator
            device.
          </p>
          <p className="text-xs">
            After this window expires, you'll need to use another recovery code
            or contact support.
          </p>
        </div>

        {/* Checkbox */}
        <div className="mb-6">
          <CheckboxField
            id="mfa-recovery-has-access"
            label="I have access to my authentication device and want to keep MFA enabled"
            checked={hasAccess}
            onChange={setHasAccess}
          />
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
          <Button
            variant="destructive"
            disabled={hasAccess || isExpired}
            loading={disabling}
            onClick={() => void handleDisable()}
          >
            Disable MFA
          </Button>
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
    </BaseDialog>
  );
}
