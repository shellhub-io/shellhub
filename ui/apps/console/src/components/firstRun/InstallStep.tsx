import type { ReactNode } from "react";
import { WindowChrome } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";

/**
 * The install command, followed by the lines the agent prints, with the code marked, so the user
 * knows what to look for before running anything. In a new namespace the command carries no
 * credential and the installer runs `shellhub-agent login`, which prints a code. In a legacy
 * namespace it carries the tenant, the device waits to be accepted, and the user runs
 * `shellhub-agent login` to print the code that accepts it. Both show the login output copied from
 * what the agent prints when it does not open a browser, which differs only in its last line: a
 * new namespace's code can be renewed while the command waits, a legacy one's expires in ten
 * minutes.
 * `tenant` is read only when `legacy` is set.
 */
export default function InstallStep({
  legacy = false,
  tenant = "",
  children,
}: {
  legacy?: boolean;
  tenant?: string;
  children?: ReactNode;
}) {
  const origin = window.location.origin;
  const installCmd = legacy
    ? `curl -sSf ${origin}/install.sh | TENANT_ID=${tenant} sh`
    : `curl -sSf ${origin}/install.sh | sh`;
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
          {legacy && (
            <>
              <span className="text-text-secondary">
                ✅ The agent is listening for connections. The device is
                pending: accept it in the console.
              </span>
              {"\n"}
              <span className="text-text-muted select-none">$ </span>
              <span className="text-accent-cyan">shellhub-agent login</span>
              {"\n"}
            </>
          )}
          {"\n"}
          <span className="text-text-secondary">
            To accept this device, open the ShellHub console and enter this
            code:
          </span>
          {"\n\n      "}
          {slot("XXXX-XXXX")}
          {"\n\n"}
          <span className="text-text-secondary">
            Enter it at {origin}/accept-device, or open the link:
          </span>
          {"\n\n  "}
          <span className="text-text-secondary">
            {origin}/accept-device?code=
          </span>
          {slot("XXXXXXXX")}
          {"\n\n"}
          <span className="text-text-secondary">
            {legacy
              ? "Waiting for acceptance... (code expires in 10 minutes)"
              : "Waiting for acceptance... (code expires in 10 minutes; a server that renews it keeps it valid while this command waits, up to an hour)"}
          </span>
        </pre>
      </WindowChrome>
      {children}
    </div>
  );
}
