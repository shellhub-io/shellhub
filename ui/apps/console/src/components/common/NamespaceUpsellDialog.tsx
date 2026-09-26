import {
  ArrowTopRightOnSquareIcon,
  CheckIcon,
  RectangleStackIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";

const PRICING_URL = "https://www.shellhub.io/pricing";

const HIGHLIGHTS = [
  "Separate namespaces for teams, environments, or customers",
  "Members and roles scoped to each namespace",
  "Devices and access policies isolated per namespace",
];

interface NamespaceUpsellDialogProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Explains why another namespace needs a paid edition, shown instead of failing the create.
 */
export default function NamespaceUpsellDialog({
  open,
  onClose,
}: NamespaceUpsellDialogProps) {
  return (
    <Modal
      layout="center"
      size="sm"
      open={open}
      onClose={onClose}
      icon={<RectangleStackIcon />}
      title="Run more than one namespace"
      description="This instance runs as a single namespace. Upgrade to keep separate workspaces, each with its own devices, members, and access policies."
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Not now
          </Button>
          <a
            href={PRICING_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-white transition-all hover:bg-primary-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
          >
            See pricing
            <ArrowTopRightOnSquareIcon className="w-4 h-4" strokeWidth={2} />
          </a>
        </>
      }
    >
      <ul className="flex flex-col gap-2.5">
        {HIGHLIGHTS.map((h) => (
          <li
            key={h}
            className="flex items-start gap-2.5 text-sm text-text-secondary"
          >
            <CheckIcon className="mt-0.5 w-4 h-4 shrink-0 text-accent-green" />
            {h}
          </li>
        ))}
      </ul>
      <p className="mt-4 text-center text-2xs text-text-muted">
        Available on Cloud and Enterprise
      </p>
    </Modal>
  );
}
