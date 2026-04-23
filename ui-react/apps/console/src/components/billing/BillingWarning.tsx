import { useNavigate } from "react-router-dom";
import ConfirmDialog from "../common/ConfirmDialog";
import { useHasPermission } from "@/hooks/useHasPermission";

interface BillingWarningProps {
  open: boolean;
  onClose: () => void;
}

export default function BillingWarning({ open, onClose }: BillingWarningProps) {
  const navigate = useNavigate();
  const canSubscribe = useHasPermission("billing:subscribe");

  const description = canSubscribe
    ? "You've reached the free-plan device limit. Subscribe to ShellHub Cloud to register unlimited devices and unlock premium features."
    : "Your namespace has reached the free-plan device limit. Ask the namespace owner to subscribe to ShellHub Cloud.";

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={() => {
        if (canSubscribe) {
          void navigate("/settings#billing");
        }
        onClose();
      }}
      title="Device limit reached"
      description={description}
      variant="warning"
      confirmLabel={canSubscribe ? "Go to billing" : "Close"}
      cancelLabel={canSubscribe ? "Not now" : "Cancel"}
    />
  );
}
