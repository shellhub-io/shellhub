import { useEffect, useState } from "react";
import {
  CheckCircleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { getFirstPendingDevice } from "@/api/devices";
import { Device } from "@/types/device";
import DistroIcon from "@/components/common/DistroIcon";
import CopyButton from "@/components/common/CopyButton";

interface WizardStep3DeviceProps {
  device: Device | null;
  onDeviceLoaded: (device: Device) => void;
}

export default function WizardStep3Device({
  device,
  onDeviceLoaded,
}: WizardStep3DeviceProps) {
  const [loading, setLoading] = useState(device === null);
  const [fetchError, setFetchError] = useState(false);
  const [expanded, setExpanded] = useState(true);

  useEffect(() => {
    if (device !== null) return;

    let cancelled = false;

    getFirstPendingDevice()
      .then((d) => {
        if (cancelled) return;
        if (d) {
          onDeviceLoaded(d);
        } else {
          setFetchError(true);
        }
      })
      .catch(() => {
        if (!cancelled) setFetchError(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [device, onDeviceLoaded]);

  return (
    <div className="py-2 flex flex-col gap-5">
      {/* Success banner */}
      <div
        role="status"
        className="flex items-center gap-3 bg-accent-green/10 border border-accent-green/25 rounded-xl px-4 py-3"
      >
        <CheckCircleIcon className="w-5 h-5 text-accent-green shrink-0" />
        <span className="text-sm font-medium text-accent-green">
          Device connection established
        </span>
      </div>

      <div>
        <h2 className="text-xl font-mono font-bold text-text-primary mb-1">
          Approve Device
        </h2>
        <p className="text-sm text-text-muted">
          A device is waiting for your approval. Review its details before
          accepting.
        </p>
      </div>

      {/* Device card */}
      {loading && (
        <div className="bg-background border border-border rounded-xl p-6 flex items-center justify-center">
          <span className="text-xs font-mono text-text-muted animate-pulse">
            Loading device&hellip;
          </span>
        </div>
      )}

      {fetchError && (
        <div role="alert" className="flex items-center gap-3 bg-accent-red/10 border border-accent-red/25 rounded-xl px-4 py-3">
          <ExclamationCircleIcon className="w-5 h-5 text-accent-red shrink-0" />
          <span className="text-sm text-accent-red">
            No pending device found. It may have already been processed.
          </span>
        </div>
      )}

      {device && (
        <article className="bg-background border border-border rounded-xl overflow-hidden">
          {/* Header row */}
          <button
            type="button"
            onClick={() => setExpanded((v) => !v)}
            className="w-full flex items-center gap-3 px-4 py-3.5 hover:bg-hover-subtle transition-colors"
          >
            <div className="w-8 h-8 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary text-lg">
              <DistroIcon id={device.info?.id ?? ""} className="text-base leading-none" />
            </div>
            <span className="flex-1 text-sm font-mono font-semibold text-text-primary text-left">
              {device.name}
            </span>
            {expanded
              ? (
                <ChevronUpIcon className="w-4 h-4 text-text-muted shrink-0" />
              )
              : (
                <ChevronDownIcon className="w-4 h-4 text-text-muted shrink-0" />
              )}
          </button>

          {/* Detail rows */}
          {expanded && (
            <dl className="border-t border-border divide-y divide-border/60">
              <DetailRow label="OS" value={device.info?.pretty_name ?? "—"} />
              <DetailRow
                label="UID"
                value={(
                  <span className="flex items-center gap-2 min-w-0">
                    <span className="font-mono text-2xs truncate">
                      {device.uid}
                    </span>
                    <CopyButton text={device.uid} size="sm" className="shrink-0" />
                  </span>
                )}
              />
              <DetailRow label="MAC" value={device.identity?.mac ?? "—"} />
              <DetailRow label="Agent" value={device.info?.version ?? "—"} />
            </dl>
          )}
        </article>
      )}
    </div>
  );
}

function DetailRow({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-center gap-4 px-4 py-2.5">
      <dt className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted w-14 shrink-0">
        {label}
      </dt>
      <dd className="text-xs text-text-secondary flex items-center flex-1 min-w-0">
        {value}
      </dd>
    </div>
  );
}
