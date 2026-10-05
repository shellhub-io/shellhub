import { useState, type ReactNode } from "react";
import { CheckCircleIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { WindowChrome } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";
import ScreenIntro from "@/components/layout/ScreenIntro";
import AuthActions from "@/components/auth/AuthActions";
import { METHODS, PAIRING_METHODS, type Method } from "./methods";

const DOCS_INSTALL_URL = "https://docs.shellhub.io/get-started/install";

/** The methods a visitor can run without an account, in the order the installer tries them. */
const OFFERED: Method[] = [
  "auto",
  "docker",
  "podman",
  "snap",
  "standalone",
  "wsl",
];

/**
 * The public install page: the command to put an agent on a machine, with no account needed to
 * read it.
 *
 * It is deliberately the Interactive path. A fleet needs a provisioning key, which needs a
 * namespace to create it in, so that half of Add Device only makes sense signed in.
 */
export default function Install() {
  const [method, setMethod] = useState<Method>("auto");

  const methods = OFFERED.map(
    (id) => METHODS.find((m) => m.id === id) as (typeof METHODS)[number],
  );
  const selected = methods.find((m) => m.id === method) ?? methods[0];
  const pairs = PAIRING_METHODS.includes(selected.id);

  const command = [
    "curl -sSf",
    `${window.location.origin}/install.sh`,
    "|",
    ...(selected.id === "auto" ? [] : [`INSTALL_METHOD=${selected.id}`]),
    ...(pairs ? [] : ["TENANT_ID=<your-tenant-id>"]),
    "sh",
  ].join(" ");

  return (
    <div>
      <ScreenIntro
        eyebrow="Install"
        title="Install ShellHub"
        lead="Run one command on the machine you want to reach. Nothing listens on it, and nothing is installed on your side."
      />

      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div
          role="tablist"
          aria-label="Installation method"
          className="flex items-center gap-1 px-3 pt-3 overflow-x-auto border-b border-border"
        >
          {methods.map((m) => {
            const isActive = m.id === selected.id;
            return (
              <button
                key={m.id}
                type="button"
                role="tab"
                aria-selected={isActive}
                onClick={() => setMethod(m.id)}
                className={cn(
                  "flex items-center gap-2 shrink-0 px-3 py-2.5 -mb-px text-sm font-medium",
                  "border-b-2 transition-colors duration-150",
                  isActive
                    ? "text-primary border-primary"
                    : "text-text-muted border-transparent hover:text-text-primary",
                )}
              >
                {m.icon}
                {m.label}
              </button>
            );
          })}
        </div>

        <div className="p-6">
          <p className="text-xs text-text-muted leading-relaxed mb-6">
            {selected.description}
          </p>

          <Step number={1} title="Run this on the device">
            <WindowChrome
              variant="terminal"
              size="sm"
              titleBarSlot={<CopyButton text={command} showLabel />}
            >
              <pre className="text-accent-cyan whitespace-pre overflow-x-auto">
                <span className="text-text-muted select-none">$ </span>
                {command}
              </pre>
            </WindowChrome>
            {!pairs && (
              <p className="text-2xs text-text-muted leading-relaxed mt-2">
                Snap is the one method that needs the namespace named up front.
                Your tenant ID is on the namespace's settings page.
              </p>
            )}
          </Step>

          <Step number={2} title="Accept it in your browser" last>
            <p className="text-xs text-text-secondary leading-relaxed">
              {pairs ? (
                <>
                  The agent prints a code and a link, and waits. Open the link,
                  or enter the code at{" "}
                  <span className="font-mono text-text-primary">
                    /accept-device
                  </span>
                  , and the device joins the namespace you pick. The code lasts
                  ten minutes.
                </>
              ) : (
                <>
                  The device enrols into the namespace you named and waits to be
                  accepted, in the activity of the provisioning key it used.
                </>
              )}
            </p>
          </Step>
        </div>

        <p className="px-6 py-4 bg-surface/40 border-t border-border text-2xs text-text-muted">
          Adding more than one? A provisioning key enrols a fleet unattended.
        </p>
      </div>

      <div className="mt-6">
        <AuthActions
          links={[
            { label: "Have a code? Enter it", to: "/accept-device" },
            { label: "Installation guide", href: DOCS_INSTALL_URL },
            { label: "Already have a namespace? Sign in", to: "/login" },
          ]}
        />
      </div>
    </div>
  );
}

function Step({
  number,
  title,
  children,
  last = false,
}: {
  number: number;
  title: string;
  children: ReactNode;
  last?: boolean;
}) {
  return (
    <div className="flex gap-4">
      <div className="flex flex-col items-center shrink-0">
        <span className="w-6 h-6 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-mono font-semibold text-primary">
          {last ? (
            <CheckCircleIcon className="w-3.5 h-3.5" strokeWidth={2} />
          ) : (
            number
          )}
        </span>
        {!last && <span className="w-px flex-1 bg-border mt-1" />}
      </div>

      <div className={cn("min-w-0 flex-1", !last && "pb-6")}>
        <p className="text-sm font-medium text-text-primary mb-2.5">{title}</p>
        {children}
      </div>
    </div>
  );
}
