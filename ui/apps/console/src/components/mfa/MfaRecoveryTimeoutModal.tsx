import { useState } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { apiErrorMessage, isSdkError } from "@/api/errors";
import CheckboxField from "@/components/common/fields/CheckboxField";
import Modal from "@/components/common/Modal";
import { useCountdown } from "@/hooks/useCountdown";
import { pluralize } from "@/utils/count";

const RECOVERY_SESSION_ENDED_MESSAGE =
  "This recovery session has expired or was already used. Sign in with another recovery code or contact support.";

interface MfaRecoveryTimeoutModalProps {
  open: boolean;
  expiresAt: number; // Unix timestamp
  onClose: () => void;
  onDisable: () => Promise<void>;
}

function minutesAndSeconds(secondsLeft: number) {
  return `${pluralize(Math.floor(secondsLeft / 60), "minute")} ${pluralize(Math.floor(secondsLeft % 60), "second")}`;
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
  const [error, setError] = useState("");
  const secondsLeft = useCountdown(expiresAt);
  const isExpired = secondsLeft === 0;

  const disableErrorMessage = (err: unknown) =>
    isSdkError(err) && (err.status === 401 || err.status === 403)
      ? RECOVERY_SESSION_ENDED_MESSAGE
      : apiErrorMessage(err);

  const handleDisable = async () => {
    setDisabling(true);
    setError("");
    try {
      await onDisable();
    } catch (err) {
      setError(disableErrorMessage(err));
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
            {secondsLeft === null
              ? ""
              : isExpired
                ? "Expired"
                : `${minutesAndSeconds(secondsLeft)} remaining`}
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
        {error && <Callout variant="error">{error}</Callout>}
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
