import { useEffect } from "react";
import { CheckCircleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useDevicePolling } from "@/hooks/useDevicePolling";
import { buildInstallCommand } from "@/utils/installCommand";
import CopyButton from "@/components/common/CopyButton";

const requirements = [
  "Linux system with curl installed",
  "Internet access on the device",
  "Docker, Podman, Snap, or standalone install",
];

interface WizardStep2InstallProps {
  onDeviceDetected: () => void;
}

export default function WizardStep2Install({
  onDeviceDetected,
}: WizardStep2InstallProps) {
  const tenant = useAuthStore((s) => s.tenant);
  const installCmd = buildInstallCommand(tenant ?? "", window.location.origin);

  const { isPolling, start } = useDevicePolling({
    onPoll: (stats) => {
      if ((stats.pending_devices ?? 0) > 0) {
        onDeviceDetected();
        return true;
      }
      return false;
    },
  });

  // Start polling as soon as this step mounts
  useEffect(() => {
    start();
  }, [start]);

  return (
    <div className="py-2 flex flex-col gap-5">
      <div>
        <h2 className="text-xl font-mono font-bold text-text-primary mb-1">
          Install the Agent
        </h2>
        <p className="text-sm text-text-muted">
          Run this command on the target device. The agent will register
          automatically.
        </p>
      </div>

      {/* Command block */}
      <div className="relative group">
        <pre className="bg-background border border-border rounded-xl p-4 pr-12 font-mono text-xs text-text-secondary leading-relaxed overflow-x-auto whitespace-pre m-0">
          <span className="text-primary/50 select-none">$ </span>
          {installCmd}
        </pre>
        <div className="absolute top-2.5 right-2.5">
          <CopyButton text={installCmd} size="md" className="bg-background border border-border/60 shadow-sm" />
        </div>
      </div>

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

      {/* Polling status */}
      <div
        role="status"
        className="flex items-center gap-3 bg-background border border-border rounded-xl px-4 py-3"
      >
        {isPolling
          ? (
            <>
              <span className="relative flex h-2.5 w-2.5 shrink-0">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-yellow opacity-60" />
                <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-accent-yellow" />
              </span>
              <span className="text-2xs font-mono text-text-muted">
                Listening for device connection&hellip;
              </span>
            </>
          )
          : (
            <>
              <span className="relative flex h-2.5 w-2.5 shrink-0">
                <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-text-muted/30" />
              </span>
              <span className="text-2xs font-mono text-text-muted">
                Waiting to start&hellip;
              </span>
            </>
          )}
      </div>
    </div>
  );
}
