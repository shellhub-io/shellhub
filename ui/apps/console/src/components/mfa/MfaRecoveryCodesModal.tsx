import { KeyIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";

interface MfaRecoveryCodesModalProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Explains why recovery codes can't be shown again: they are readable only once, during MFA setup,
 * and the way to new ones is to turn MFA off and on.
 */
export default function MfaRecoveryCodesModal({
  open,
  onClose,
}: MfaRecoveryCodesModalProps) {
  return (
    <Modal
      layout="center"
      open={open}
      onClose={onClose}
      icon={<KeyIcon />}
      iconColor="yellow"
      title="Recovery codes"
      description="Recovery codes are shown only once, during MFA setup. To get new ones, turn MFA off and on again."
      footer={
        <Button variant="ghost" onClick={onClose}>
          Close
        </Button>
      }
    />
  );
}
