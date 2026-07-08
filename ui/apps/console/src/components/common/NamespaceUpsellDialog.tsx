import {
  ArrowTopRightOnSquareIcon,
  CheckIcon,
  RectangleStackIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "@/components/common/BaseDialog";

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

export default function NamespaceUpsellDialog({
  open,
  onClose,
}: NamespaceUpsellDialogProps) {
  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-labelledby="namespace-upsell-title"
    >
      <button
        type="button"
        aria-label="Close"
        onClick={onClose}
        className="absolute right-4 top-4 rounded-md p-1.5 text-text-muted transition-colors hover:text-text-secondary"
      >
        <XMarkIcon className="w-5 h-5" />
      </button>

      <div className="bg-gradient-to-b from-primary/[0.1] to-transparent px-7 pb-2 pt-8">
        <span className="flex h-12 w-12 items-center justify-center rounded-xl border border-primary/25 bg-primary/10 text-primary shadow-lg shadow-primary/10">
          <RectangleStackIcon className="w-6 h-6" />
        </span>
        <p className="mt-4 font-mono text-2xs font-semibold uppercase tracking-wider text-accent-yellow">
          Premium
        </p>
        <h2
          id="namespace-upsell-title"
          className="mt-1 text-lg font-semibold text-balance text-text-primary"
        >
          Run more than one namespace
        </h2>
      </div>

      <div className="px-7 pb-6 pt-2">
        <p className="text-sm text-text-secondary">
          This instance runs as a single namespace. Upgrade to keep separate
          workspaces, each with its own devices, members, and access policies.
        </p>
        <ul className="mt-4 flex flex-col gap-2.5">
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
      </div>

      <div className="flex items-center justify-between gap-3 border-t border-border px-7 py-4">
        <p className="text-2xs text-text-muted">
          Available on Cloud and Enterprise
        </p>
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={onClose}
            className="text-sm font-medium text-text-secondary transition-colors hover:text-text-primary"
          >
            Not now
          </button>
          <a
            href={PRICING_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-semibold text-white shadow-lg shadow-primary/20 transition-all hover:bg-primary-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
          >
            See pricing
            <ArrowTopRightOnSquareIcon className="w-4 h-4" strokeWidth={2} />
          </a>
        </div>
      </div>
    </BaseDialog>
  );
}
