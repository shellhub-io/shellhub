import type { ReactNode } from "react";
import {
  ArrowTopRightOnSquareIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "@/components/common/Modal";
import { PRICING_URL } from "@/edition";

interface UpsellDialogProps {
  open: boolean;
  onClose: () => void;
  icon: ReactNode;
  title: string;
  description: string;
  highlights: readonly string[];
}

/**
 * Explains what a paid edition adds for one feature, with its highlights and a link to the
 * pricing page, shown in place of the feature on community.
 */
export default function UpsellDialog({
  open,
  onClose,
  icon,
  title,
  description,
  highlights,
}: UpsellDialogProps) {
  return (
    <Modal
      layout="center"
      size="sm"
      open={open}
      onClose={onClose}
      icon={icon}
      title={title}
      description={description}
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
        {highlights.map((h) => (
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
