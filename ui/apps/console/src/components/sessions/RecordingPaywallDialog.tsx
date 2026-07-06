import {
  ArrowTopRightOnSquareIcon,
  CheckIcon,
  VideoCameraIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "@/components/common/BaseDialog";

const PRICING_URL = "https://www.shellhub.io/pricing";

const HIGHLIGHTS = [
  "Recorded on the server, independent of anyone's browser",
  "Retained centrally for audit and compliance",
  "Replay any session later, keystroke by keystroke",
];

interface RecordingPaywallDialogProps {
  open: boolean;
  onClose: () => void;
}

export default function RecordingPaywallDialog({
  open,
  onClose,
}: RecordingPaywallDialogProps) {
  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-labelledby="recording-paywall-title"
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
          <VideoCameraIcon className="w-6 h-6" />
        </span>
        <p className="mt-4 font-mono text-2xs font-semibold uppercase tracking-wider text-accent-yellow">
          Premium
        </p>
        <h2
          id="recording-paywall-title"
          className="mt-1 text-lg font-semibold text-balance text-text-primary"
        >
          Record and replay every session
        </h2>
      </div>

      <div className="px-7 pb-6 pt-2">
        <p className="text-sm text-text-secondary">
          This session was not recorded. Advanced Session Recording captures
          every terminal session on the server, so you never depend on a browser
          copy that only lives on one machine.
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
