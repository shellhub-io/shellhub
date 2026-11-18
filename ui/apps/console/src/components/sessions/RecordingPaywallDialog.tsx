import { VideoCameraIcon } from "@heroicons/react/24/outline";
import UpsellDialog from "@/components/common/UpsellDialog";

const HIGHLIGHTS = [
  "Recorded on the server, independent of anyone's browser",
  "Retained centrally for audit and compliance",
  "Replay any session later, keystroke by keystroke",
];

interface RecordingPaywallDialogProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Explains that session recording needs a paid plan.
 */
export default function RecordingPaywallDialog({
  open,
  onClose,
}: RecordingPaywallDialogProps) {
  return (
    <UpsellDialog
      open={open}
      onClose={onClose}
      icon={<VideoCameraIcon />}
      title="Record and replay every session"
      description="This session was not recorded. Advanced Session Recording captures every terminal session on the server, so you never depend on a browser copy that only lives on one machine."
      highlights={HIGHLIGHTS}
    />
  );
}
