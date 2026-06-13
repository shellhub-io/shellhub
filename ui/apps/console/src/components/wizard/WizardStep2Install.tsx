import { useEffect } from "react";
import { CheckCircleIcon } from "@heroicons/react/24/outline";
import { Button, WindowChrome } from "@shellhub/design-system/primitives";
import { useAuthStore } from "@/stores/authStore";
import { useDevicePairingEnrollment } from "@/hooks/useDevicePairingEnrollment";
import { buildInstallCommand } from "@/utils/installCommand";
import { formatCountdown } from "@/utils/date";
import CopyButton from "@/components/common/CopyButton";

const requirements = [
  "Linux system with curl installed",
  "Internet access on the device",
  "Docker, Podman, Snap, or standalone install",
];

interface WizardStep2InstallProps {
  onConnected: (device: { uid: string; name: string }) => void;
}

export default function WizardStep2Install({
  onConnected,
}: WizardStep2InstallProps) {
  const tenant = useAuthStore((s) => s.tenant);
  const enrollment = useDevicePairingEnrollment(true, tenant ?? "");

  const commandReady = Boolean(enrollment.code);
  const installCmd = commandReady
    ? buildInstallCommand(`CODE=${enrollment.code}`, window.location.origin)
    : "";

  // The device is accepted automatically the moment it claims the code, so move
  // straight to the finish step — no manual approval.
  useEffect(() => {
    if (enrollment.phase === "connected" && enrollment.device) {
      onConnected(enrollment.device);
    }
  }, [enrollment.phase, enrollment.device, onConnected]);

  return (
    <div className="py-2 flex flex-col gap-5">
      <div>
        <h2 className="text-xl font-mono font-bold text-text-primary mb-1">
          Install the Agent
        </h2>
        <p className="text-sm text-text-muted">
          Run this command on the target device. It joins this namespace and
          shows up here automatically.
        </p>
      </div>

      {/* Command block */}
      <WindowChrome
        variant="terminal"
        size="sm"
        titleBarSlot={
          commandReady ? <CopyButton text={installCmd} showLabel /> : null
        }
      >
        <pre className="text-accent-cyan whitespace-pre-wrap break-all">
          <span className="text-text-muted select-none">$ </span>
          {commandReady ? (
            installCmd
          ) : (
            <span className="text-text-muted">
              Preparing your install command…
            </span>
          )}
        </pre>
      </WindowChrome>

      {/* Requirements */}
      <div>
        <h3 className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
          Requirements
        </h3>
        <ul className="flex flex-col gap-2">
          {requirements.map((req) => (
            <li key={req} className="flex items-start gap-2.5">
              <CheckCircleIcon className="w-4 h-4 text-accent-green shrink-0 mt-px" />
              <span className="text-xs text-text-secondary">{req}</span>
            </li>
          ))}
        </ul>
      </div>

      {/* Live status */}
      <div
        role="status"
        className="flex items-center gap-3 bg-background border border-border rounded-xl px-4 py-3"
      >
        {enrollment.phase === "expired" ? (
          <>
            <span className="relative inline-flex h-2.5 w-2.5 shrink-0 rounded-full bg-text-muted/40" />
            <span className="flex-1 text-2xs font-mono text-text-muted">
              The code expired.
            </span>
            <Button variant="ghost" size="sm" onClick={enrollment.regenerate}>
              New code
            </Button>
          </>
        ) : enrollment.phase === "error" ? (
          <>
            <span className="relative inline-flex h-2.5 w-2.5 shrink-0 rounded-full bg-accent-red/60" />
            <span className="flex-1 text-2xs font-mono text-text-muted">
              Couldn&apos;t prepare a code.
            </span>
            <Button variant="ghost" size="sm" onClick={enrollment.regenerate}>
              Retry
            </Button>
          </>
        ) : (
          <>
            <span className="relative flex h-2.5 w-2.5 shrink-0">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-yellow opacity-60" />
              <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-accent-yellow" />
            </span>
            <span className="flex-1 text-2xs font-mono text-text-muted">
              Listening for device connection&hellip;
              {enrollment.secondsLeft > 0
                ? ` code expires in ${formatCountdown(enrollment.secondsLeft)}`
                : ""}
            </span>
            <Button variant="ghost" size="sm" onClick={enrollment.regenerate}>
              New code
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
