import { ChatBubbleLeftRightIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";

const PRICING_URL = "https://www.shellhub.io/pricing";
const DOCS_URL = "https://docs.shellhub.io/";

interface SupportPaywallDialogProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Explains that chat support needs a paid plan, and where to get one. Upgrading opens the pricing
 * page and calls onClose as well, so onClose means the dialog is done, not that the user declined.
 */
export default function SupportPaywallDialog({
  open,
  onClose,
}: SupportPaywallDialogProps) {
  const handleUpgrade = () => {
    window.open(PRICING_URL, "_blank", "noopener,noreferrer");
    onClose();
  };

  return (
    <Modal
      layout="center"
      size="sm"
      open={open}
      onClose={onClose}
      icon={<ChatBubbleLeftRightIcon />}
      title="Upgrade to access chat support"
      description="Chat with our team directly, with priority responses, on a paid plan."
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Not now
          </Button>
          <Button onClick={handleUpgrade}>Upgrade</Button>
        </>
      }
    >
      <p className="text-center text-sm text-text-muted">
        You can still browse our{" "}
        <a
          href={DOCS_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary hover:text-primary-300 underline underline-offset-2"
        >
          documentation
        </a>{" "}
        to find answers and troubleshoot on your own.
      </p>
    </Modal>
  );
}
