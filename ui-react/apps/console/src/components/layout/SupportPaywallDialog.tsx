import ConfirmDialog from "@/components/common/ConfirmDialog";

const PRICING_URL = "https://www.shellhub.io/pricing";
const DOCS_URL = "https://docs.shellhub.io/";

interface SupportPaywallDialogProps {
  open: boolean;
  onClose: () => void;
}

export default function SupportPaywallDialog({
  open,
  onClose,
}: SupportPaywallDialogProps) {
  const handleUpgrade = () => {
    window.open(PRICING_URL, "_blank", "noopener,noreferrer");
    onClose();
  };

  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={handleUpgrade}
      title="Upgrade to access chat support"
      variant="primary"
      confirmLabel="Upgrade"
      cancelLabel="Close"
      description={
        <div className="space-y-3">
          <p className="text-sm text-text-secondary">
            Get real-time assistance from our team with priority responses. Skip
            the documentation hunt — upgrade now and unlock direct chat support.
          </p>
          <p className="text-sm text-text-muted">
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
        </div>
      }
    />
  );
}
