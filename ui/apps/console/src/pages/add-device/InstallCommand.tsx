import { useState, type ReactNode } from "react";
import {
  ArrowTopRightOnSquareIcon,
  BookOpenIcon,
  ChevronRightIcon,
} from "@heroicons/react/24/outline";
import { Button, Card, WindowChrome } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import CopyButton from "@/components/common/CopyButton";
import InputField from "@/components/common/fields/InputField";
import NumericInput from "@/components/common/fields/NumericInput";
import { METHODS, type Method } from "@/pages/install/methods";

const OPTIONAL = (
  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
    (optional)
  </span>
);

/**
 * The install command for a method, with the credential the caller puts on it, the agent's
 * optional flags under Advanced options, and a note on what happens once it runs. A manual method
 * gets its platform guide instead, since the installer does not run it.
 */
export default function InstallCommand({
  method,
  credential,
  outcome,
}: {
  method: Method;
  credential?: string;
  outcome: ReactNode;
}) {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [hostname, setHostname] = useState("");
  const [identity, setIdentity] = useState("");
  const [keepaliveInterval, setKeepaliveInterval] = useState("");
  const keepaliveIntervalError =
    keepaliveInterval && parseInt(keepaliveInterval, 10) < 1
      ? "Interval must be a positive number"
      : "";

  const info = METHODS.find((m) => m.id === method)!;

  if (info.manual) {
    return (
      <Card className="p-5">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center shrink-0">
            <BookOpenIcon className="w-4.5 h-4.5 text-accent-yellow" />
          </div>
          <div>
            <p className="text-sm text-text-primary font-medium mb-1">
              Manual installation required
            </p>
            <p className="text-2xs text-text-muted leading-relaxed mb-3">
              {info.label} is set up with the platform&apos;s own toolchain. The
              guide walks through it.
            </p>
            <Button
              variant="warningSoft"
              as="a"
              size="sm"
              href={info.docsUrl}
              target="_blank"
              rel="noopener noreferrer"
              iconRight={
                <ArrowTopRightOnSquareIcon
                  className="w-3 h-3"
                  strokeWidth={2}
                />
              }
            >
              View {info.label} guide
            </Button>
          </div>
        </div>
      </Card>
    );
  }

  const parts = ["curl -sSf", `${window.location.origin}/install.sh`, "|"];
  if (method !== "auto") parts.push(`INSTALL_METHOD=${method}`);
  if (credential) parts.push(credential);
  if (hostname.trim()) parts.push(`PREFERRED_HOSTNAME=${hostname.trim()}`);
  if (identity.trim()) parts.push(`PREFERRED_IDENTITY=${identity.trim()}`);
  if (keepaliveInterval.trim() && !keepaliveIntervalError)
    parts.push(`KEEPALIVE_INTERVAL=${parseInt(keepaliveInterval, 10)}`);
  parts.push("sh");
  const command = parts.join(" ");

  return (
    <div className="space-y-3">
      <WindowChrome
        variant="terminal"
        size="sm"
        titleBarSlot={<CopyButton text={command} showLabel />}
      >
        <pre className="text-accent-cyan whitespace-pre-wrap break-all">
          <span className="text-text-muted select-none">$ </span>
          {command}
        </pre>
      </WindowChrome>

      <button
        type="button"
        onClick={() => setShowAdvanced(!showAdvanced)}
        aria-expanded={showAdvanced}
        className="flex items-center gap-1.5 text-2xs font-mono text-text-muted hover:text-text-secondary transition-colors"
      >
        <ChevronRightIcon
          className={cn(
            "w-3 h-3 transition-transform",
            showAdvanced && "rotate-90",
          )}
          strokeWidth={2}
        />
        Advanced options
      </button>

      {showAdvanced && (
        <Card className="p-4 space-y-4 animate-fade-in">
          <InputField
            id="add-device-hostname"
            label="Preferred Hostname"
            labelAdornment={OPTIONAL}
            value={hostname}
            onChange={setHostname}
            placeholder="my-device"
            hint="Override the device hostname reported to ShellHub."
          />
          <InputField
            id="add-device-identity"
            label="Preferred Identity"
            labelAdornment={OPTIONAL}
            value={identity}
            onChange={setIdentity}
            placeholder="server-01"
            hint="Set a custom identity string for the device."
          />
          <NumericInput
            id="add-device-keepalive"
            label="Keep Alive Interval"
            labelAdornment={OPTIONAL}
            value={keepaliveInterval}
            onChange={setKeepaliveInterval}
            placeholder="30"
            hint="Interval in seconds between keep-alive messages sent by the agent. Defaults to 30."
            error={keepaliveIntervalError || undefined}
          />
        </Card>
      )}

      <p className="text-xs text-text-secondary leading-relaxed">{outcome}</p>
    </div>
  );
}
