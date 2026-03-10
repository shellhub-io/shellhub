import { useId } from "react";
import {
  XMarkIcon,
  BookOpenIcon,
  FolderPlusIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "./BaseDialog";
import CopyButton from "./CopyButton";

const COMMAND = "./bin/cli namespace create <namespace> <owner>";

const rules = [
  "3–30 characters",
  "Lowercase letters, numbers, and hyphens only",
  "Cannot begin or end with a hyphen",
];

interface CreateNamespaceDialogProps {
  open: boolean;
  onClose: () => void;
}

export default function CreateNamespaceDialog({
  open,
  onClose,
}: CreateNamespaceDialogProps) {
  const autoId = useId();
  const titleId = `create-ns-title-${autoId}`;
  const descriptionId = `create-ns-description-${autoId}`;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="lg"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
    >
      {/* Header */}
      <header className="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border shrink-0">
        <div className="flex items-center gap-3">
          <span className="w-8 h-8 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
            <FolderPlusIcon className="w-4 h-4" />
          </span>
          <h2
            id={titleId}
            className="text-sm font-semibold text-text-primary"
          >
            Create a Namespace
          </h2>
        </div>

        <button
          onClick={onClose}
          className="p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all"
          aria-label="Close dialog"
        >
          <XMarkIcon className="w-4 h-4" />
        </button>
      </header>

      {/* Body */}
      <div className="px-6 py-5 space-y-5">
        <p
          id={descriptionId}
          className="text-sm text-text-muted leading-relaxed"
        >
          Community Edition uses the CLI to manage namespaces. Run this
          command on your server:
        </p>

        {/* Command block */}
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <div className="flex items-center justify-between px-4 py-2.5 border-b border-border bg-surface/50">
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-accent-red/60" />
              <span className="w-2.5 h-2.5 rounded-full bg-accent-yellow/60" />
              <span className="w-2.5 h-2.5 rounded-full bg-accent-green/60" />
            </div>
            <span className="text-2xs font-mono text-text-muted/50">terminal</span>
            <CopyButton text={COMMAND} showLabel />
          </div>
          <div className="p-4 overflow-x-auto">
            <pre className="text-xs font-mono text-accent-cyan leading-relaxed whitespace-pre m-0">
              <span className="text-text-muted select-none">$ </span>{COMMAND}
            </pre>
          </div>
        </div>

        {/* Name rules */}
        <div>
          <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
            Naming rules
          </p>
          <ul className="space-y-1.5" aria-label="Namespace naming rules">
            {rules.map((rule) => (
              <li
                key={rule}
                className="flex items-start gap-2 text-xs text-text-muted"
              >
                <span
                  className="w-1 h-1 rounded-full bg-border-light mt-1.5 shrink-0"
                  aria-hidden="true"
                />
                {rule}
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* Footer */}
      <footer className="px-6 py-4 border-t border-border flex items-center justify-between shrink-0">
        <a
          href="https://docs.shellhub.io/self-hosted/administration"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          <BookOpenIcon className="w-3.5 h-3.5" />
          Administration Guide
        </a>

        <button
          onClick={onClose}
          className="px-4 py-2 rounded-lg text-xs font-medium text-text-secondary hover:text-text-primary hover:bg-hover-medium transition-all"
        >
          Close
        </button>
      </footer>
    </BaseDialog>
  );
}
