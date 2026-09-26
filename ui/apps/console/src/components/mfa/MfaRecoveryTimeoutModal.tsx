import { useState } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useCountdown } from "@/hooks/useCountdown";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";

interface MfaRecoveryTimeoutModalProps {
  open: boolean;
  expiresAt: number; // Unix timestamp
  onClose: () => void;
  onDisable: () => Promise<void>;
}

/**
 * Counts down the window in which an MFA reset can be completed, and says what happens when it
 * closes — the request expires server-side, so the countdown is a fact rather than a nudge.
 */
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
    <Modal
      open={open}
      onClose={onClose}
      canClose={() => false}
      size="sm"
      icon={<ExclamationTriangleIcon />}
      iconColor="yellow"
      title="Recovery window active"
      description={
        <>
          <span className="font-mono text-accent-yellow">
            {isExpired ? "Expired" : `${timeLeft} remaining`}
          </span>
          . You signed in with a recovery code, so you can turn MFA off now if
          you lost your authenticator.
        </>
      }
      footer={
        <Button
          variant="destructive"
          disabled={hasAccess || isExpired}
          loading={disabling}
          onClick={() => void handleDisable()}
        >
          Disable MFA
        </Button>
      }
    >
      <div className="space-y-4">
        <p className="text-xs text-text-muted leading-relaxed">
          Once the window closes, you'll need another recovery code or to
          contact support.
        </p>
        <CheckboxField
          id="mfa-recovery-has-access"
          label="I have access to my authentication device and want to keep MFA enabled"
          checked={hasAccess}
          onChange={setHasAccess}
        />
      </div>
    </Modal>
  );
}
