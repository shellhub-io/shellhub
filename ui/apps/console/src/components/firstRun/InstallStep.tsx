import type { ReactNode } from "react";
import { WindowChrome } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";

/**
 * The install command, followed by the lines the agent prints when it has no namespace yet, with
 * the pairing code marked, so the user knows what to look for before running anything. The copy
 * of those lines follows the agent's own log output.
 */
export default function InstallStep({ children }: { children?: ReactNode }) {
  const origin = window.location.origin;
  const installCmd = `curl -sSf ${origin}/install.sh | sh`;
  const slot = (text: string) => (
    <span className="text-accent-yellow bg-accent-yellow/10 rounded px-1">
      {text}
    </span>
  );

  return (
    <div className="flex flex-col gap-5">
      <WindowChrome
        variant="terminal"
        size="sm"
        title="Run on the device"
        titleBarSlot={<CopyButton text={installCmd} showLabel />}
      >
        <pre className="text-xs leading-relaxed whitespace-pre-wrap [overflow-wrap:anywhere]">
          <span className="text-text-muted select-none">$ </span>
          <span className="text-accent-cyan">{installCmd}</span>
          {"\n"}
          <span className="text-text-muted">…</span>
          {"\n"}
          <span className="text-text-secondary">
            This device is not enrolled in any namespace yet.
            {"\n"}To pair it, open {origin}/accept-device and enter code{" "}
          </span>
          {slot("XXXX-XXXX")}
          {"\n"}
          <span className="text-text-secondary">
            Or open the link directly: {origin}/accept-device?code=
          </span>
          {slot("XXXXXXXX")}
          {"\n"}
          <span className="text-text-secondary">
            Waiting for acceptance... (code expires in 10 minutes)
          </span>
        </pre>
      </WindowChrome>
      {children}
    </div>
  );
}
