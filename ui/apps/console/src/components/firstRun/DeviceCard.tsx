import type { ReactNode } from "react";
import type { DeviceIdentity, DeviceInfo } from "@/client";
import { formatPairingCode } from "@/utils/pairingCode";

interface DeviceCardProps {
  device: {
    name?: string;
    info?: DeviceInfo | null;
    identity?: DeviceIdentity;
  };
  code?: string;
  status: ReactNode;
}

/**
 * A device as the first run shows it: its name and system, what identifies it (the pairing code
 * too, when given), and where it stands. The pair step shows the device behind a code; the shell
 * step shows it once paired.
 */
export default function DeviceCard({ device, code, status }: DeviceCardProps) {
  const system = [device.info?.pretty_name, device.info?.arch]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="bg-card border border-border rounded-xl">
      <div className="flex items-center gap-3.5 p-4">
        <div className="min-w-0 flex-1">
          <p className="font-mono text-sm font-semibold text-text-primary truncate">
            {device.name}
          </p>
          {system && (
            <p className="text-xs text-text-muted truncate">{system}</p>
          )}
        </div>
        <div className="shrink-0">{status}</div>
      </div>
      <dl
        className={
          code
            ? "grid grid-cols-1 sm:grid-cols-3 border-t border-border"
            : "grid grid-cols-1 sm:grid-cols-2 border-t border-border"
        }
      >
        <Fact label="MAC" value={device.identity?.mac} />
        <Fact label="Agent" value={device.info?.version} />
        {code && <Fact label="Code" value={formatPairingCode(code)} />}
      </dl>
    </div>
  );
}

function Fact({ label, value }: { label: string; value?: string }) {
  return (
    <div className="px-4 py-2.5 border-b sm:border-b-0 sm:border-r last:border-0 border-border">
      <dt className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
        {label}
      </dt>
      <dd className="font-mono text-xs text-text-secondary">{value || "—"}</dd>
    </div>
  );
}
