import { useState, type ReactNode } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";

/**
 * Asks before a dirty form in a modal is thrown away. Wire requestClose to every close the user
 * starts (the modal's onClose, which covers the close button, Escape and the backdrop, and the
 * Cancel button) and render prompt beside the modal. Closing a form that is not dirty goes
 * straight to onClose; a caller that closes the modal itself, as it does after a save, calls
 * onClose directly and is never asked.
 */
export function useDiscardGuard({
  open,
  dirty,
  onClose,
}: {
  open: boolean;
  dirty: boolean;
  onClose: () => void;
}): { requestClose: () => void; prompt: ReactNode } {
  const [confirming, setConfirming] = useState(false);
  useResetOnOpen(open, () => setConfirming(false));

  const requestClose = () => {
    if (dirty) setConfirming(true);
    else onClose();
  };

  const prompt = (
    <ConfirmDialog
      open={confirming}
      onClose={() => setConfirming(false)}
      onConfirm={() => {
        setConfirming(false);
        onClose();
      }}
      icon={<ExclamationTriangleIcon />}
      title="Discard changes?"
      description="What you typed in this form will be lost."
      confirmLabel="Discard changes"
      cancelLabel="Keep editing"
    />
  );

  return { requestClose, prompt };
}
