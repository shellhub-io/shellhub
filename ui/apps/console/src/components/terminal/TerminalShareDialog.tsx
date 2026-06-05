import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import {
  XMarkIcon,
  ShareIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";
import { useTerminalStore } from "@/stores/terminalStore";
import type { TerminalSession } from "@/stores/terminalStore";

const DURATIONS: { label: string; ttl: number }[] = [
  { label: "Default", ttl: 0 },
  { label: "30 minutes", ttl: 30 * 60 },
  { label: "1 hour", ttl: 60 * 60 },
  { label: "4 hours", ttl: 4 * 60 * 60 },
  { label: "No limit", ttl: -1 },
];

export default function TerminalShareDialog({
  session,
  open,
  onClose,
}: {
  session: TerminalSession;
  open: boolean;
  onClose: () => void;
}) {
  const requestShare = useTerminalStore((s) => s.requestShare);
  const token = useTerminalStore(
    (s) => s.sessions.find((ss) => ss.id === session.id)?.shareToken,
  );

  const [name, setName] = useState("");
  const [writable, setWritable] = useState(false);
  const [ttl, setTtl] = useState(0);
  const [copied, setCopied] = useState(false);

  // Reset the form each time the dialog opens.
  useEffect(() => {
    if (open) {
      setName("");
      setWritable(false);
      setTtl(0);
      setCopied(false);
    }
  }, [open]);

  if (!open) return null;

  const url = token ? `${window.location.origin}/share/${token}` : "";

  const handleCreate = () => {
    requestShare(session.id, { name: name.trim(), writable, ttl });
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard unavailable — ignore.
    }
  };

  return createPortal(
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-xl border border-border bg-surface shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
          <div className="flex items-center gap-2 text-text-primary">
            <ShareIcon className="h-4 w-4 text-primary" />
            <span className="text-sm font-semibold">Share terminal</span>
          </div>
          <button
            onClick={onClose}
            className="rounded-md p-1 text-text-muted transition-colors hover:bg-card hover:text-text-primary"
          >
            <XMarkIcon className="h-4 w-4" />
          </button>
        </div>

        {!token ? (
          <div className="space-y-4 p-5">
            <p className="text-xs text-text-secondary">
              Create a public link to this live session. Anyone with the link can watch — no sign-in
              required.
            </p>

            <div className="space-y-1.5">
              <label className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
                Name (optional)
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Debugging the deploy"
                className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-text-primary placeholder:text-text-muted/60 focus:border-primary focus:outline-none"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
                Time limit
              </label>
              <div className="flex flex-wrap gap-1.5">
                {DURATIONS.map((d) => (
                  <button
                    key={d.label}
                    onClick={() => setTtl(d.ttl)}
                    className={`rounded-md border px-2.5 py-1 text-2xs font-semibold transition-all ${
                      ttl === d.ttl
                        ? "border-primary/30 bg-primary/10 text-primary"
                        : "border-border text-text-muted hover:text-text-secondary"
                    }`}
                  >
                    {d.label}
                  </button>
                ))}
              </div>
            </div>

            <label className="flex items-center justify-between rounded-md border border-border bg-card px-3 py-2.5">
              <span className="flex flex-col">
                <span className="text-sm text-text-primary">Allow typing</span>
                <span className="text-2xs text-text-muted">
                  Guests can type into this terminal (collaborative)
                </span>
              </span>
              <button
                role="switch"
                aria-checked={writable}
                onClick={() => setWritable((w) => !w)}
                className={`relative h-5 w-9 shrink-0 rounded-full transition-colors ${
                  writable ? "bg-primary" : "bg-border-light"
                }`}
              >
                <span
                  className={`absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform ${
                    writable ? "translate-x-4" : "translate-x-0.5"
                  }`}
                />
              </button>
            </label>

            <button
              onClick={handleCreate}
              className="w-full rounded-md bg-primary px-3 py-2 text-sm font-semibold text-white transition-all hover:brightness-110"
            >
              Create link
            </button>
          </div>
        ) : (
          <div className="space-y-4 p-5">
            <div className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]" />
              <span className="text-sm text-text-primary">
                Sharing live{" "}
                <span className="text-text-muted">
                  ({writable ? "collaborative" : "read-only"})
                </span>
              </span>
            </div>

            <div className="flex items-center gap-2 rounded-md border border-border bg-card px-3 py-2">
              <code className="flex-1 truncate text-xs font-mono text-text-secondary">
                {url}
              </code>
              <button
                onClick={() => void handleCopy()}
                title="Copy link"
                className={`inline-flex items-center gap-1 rounded-md border px-2 py-1 text-2xs font-semibold transition-all ${
                  copied
                    ? "border-accent-green/20 bg-accent-green/10 text-accent-green"
                    : "border-primary/20 bg-primary/10 text-primary hover:bg-primary/20"
                }`}
              >
                {copied ? (
                  <CheckIcon className="h-3 w-3" />
                ) : (
                  <ClipboardDocumentIcon className="h-3 w-3" />
                )}
                {copied ? "Copied" : "Copy"}
              </button>
              <a
                href={url}
                target="_blank"
                rel="noreferrer"
                title="Open in new tab"
                className="inline-flex items-center rounded-md border border-border px-2 py-1 text-text-muted transition-all hover:text-text-primary"
              >
                <ArrowTopRightOnSquareIcon className="h-3 w-3" />
              </a>
            </div>

            <p className="text-2xs text-text-muted">
              The share ends when you close this terminal. Manage active shares under{" "}
              <span className="text-text-secondary">Shared Terminals</span>.
            </p>
          </div>
        )}
      </div>
    </div>,
    document.body,
  );
}
