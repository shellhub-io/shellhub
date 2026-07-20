import { ReactNode } from "react";
import { CheckCircleIcon } from "@heroicons/react/24/outline";
import { WindowChrome } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";

/**
 * The install face of step 1: the code-less command, requirements, and a live
 * status. It's presentational — the accepted-device polling lives in
 * WizardAcceptedWatcher (mounted across both faces of step 1), so the link path
 * keeps working even after the user switches to the code-entry face.
 */
export default function WizardStepInstall() {
  // Code-less install: a clean command with no credential. The agent boots
  // tenant-less and prints an accept link (and a code), then waits.
  const installCmd = `curl -sSf ${window.location.origin}/install.sh | sh`;

  const requirements: { key: string; label: ReactNode }[] = [
    { key: "curl", label: "Linux system with curl installed" },
    {
      key: "reach",
      label: (
        <>
          Can reach{" "}
          <code className="font-mono text-text-primary bg-background border border-border rounded px-1 py-px">
            {window.location.host}
          </code>
        </>
      ),
    },
    { key: "runtime", label: "Docker or Podman" },
  ];

  return (
    <div className="py-2 flex flex-col gap-5">
      <div>
        <h2 className="text-lg font-semibold text-text-primary mb-1">
          Let&apos;s connect your first device
        </h2>
        <p className="text-sm text-text-muted">
          Run this command on the target device. Once accepted, it shows up
          here.
        </p>
      </div>

      <WindowChrome
        variant="terminal"
        size="sm"
        titleBarSlot={<CopyButton text={installCmd} showLabel />}
      >
        <pre className="text-accent-cyan whitespace-pre-wrap break-all">
          <span className="text-text-muted select-none">$ </span>
          {installCmd}
        </pre>
      </WindowChrome>

      <div>
        <h3 className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
          Requirements
        </h3>
        <ul className="flex flex-col gap-2">
          {requirements.map((req) => (
            <li key={req.key} className="flex items-start gap-2.5">
              <CheckCircleIcon className="w-4 h-4 text-accent-green shrink-0 mt-px" />
              <span className="text-xs text-text-secondary">{req.label}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
