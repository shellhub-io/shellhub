import { Link } from "react-router-dom";
import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useDeviceCapacity } from "@/hooks/useDeviceCapacity";
import { formatCount } from "@/utils/count";
import { formatLicenseTimestamp, type DeviceCapacity } from "@/utils/license";
import { LABEL_BASE } from "@/utils/styles";
import { Panel, PanelError, Row } from "./Panel";

const BAR_COLOR: Record<DeviceCapacity["state"], string> = {
  unlimited: "bg-primary",
  ok: "bg-primary",
  approaching: "bg-accent-yellow",
  over: "bg-accent-red",
};

function DeviceUsage({ capacity }: { capacity: DeviceCapacity }) {
  return (
    <div className="px-5 pb-5">
      <p className={cn(LABEL_BASE, "mb-2")}>Devices</p>
      <p className="flex items-baseline gap-2">
        <span className="text-4xl font-mono font-bold tabular-nums text-text-primary">
          {formatCount(capacity.used)}
        </span>
        <span className="text-sm text-text-muted">
          {capacity.state === "unlimited"
            ? "accepted, unlimited"
            : `of ${formatCount(capacity.limit)} accepted`}
        </span>
      </p>
      {capacity.state !== "unlimited" && capacity.limit > 0 && (
        <div
          role="progressbar"
          aria-label="Licensed devices in use"
          aria-valuemin={0}
          aria-valuemax={capacity.limit}
          aria-valuenow={Math.min(capacity.used, capacity.limit)}
          aria-valuetext={`${formatCount(capacity.used)} of ${formatCount(capacity.limit)} devices`}
          className="mt-4 h-2 rounded-full bg-border overflow-hidden"
        >
          <div
            className={cn(
              "h-full rounded-full transition-all",
              BAR_COLOR[capacity.state],
            )}
            style={{
              width: `${Math.min(100, (capacity.used / capacity.limit) * 100)}%`,
            }}
          />
        </div>
      )}
    </div>
  );
}

/**
 * The installed licence: accepted devices against its limit, when it expires and who it was
 * issued to. A licence with no expiry reads as Never.
 */
export default function LicensePanel() {
  const {
    license: installedLicense,
    capacity,
    isLoading,
    failed,
  } = useDeviceCapacity();
  const customer = installedLicense?.customer;
  const issuedTo = customer?.company || customer?.name;

  return (
    <Panel
      title="License"
      action={
        <Link
          to="/admin/license"
          className="flex items-center gap-1 text-xs text-text-muted hover:text-primary"
        >
          Details
          <ArrowRightIcon className="w-3 h-3" />
        </Link>
      }
    >
      {failed === "license" && (
        <PanelError>Couldn&apos;t load the license.</PanelError>
      )}
      {failed === "stats" && (
        <PanelError>Couldn&apos;t load the accepted device count.</PanelError>
      )}
      {isLoading && !installedLicense && (
        <p role="status" className="px-5 pb-5 text-sm text-text-muted">
          Loading the license…
        </p>
      )}
      {capacity && <DeviceUsage capacity={capacity} />}
      {installedLicense && (
        <>
          <Row label="Expires">
            {installedLicense.expires_at > 0
              ? formatLicenseTimestamp(installedLicense.expires_at)
              : "Never"}
          </Row>
          {issuedTo && <Row label="Issued to">{issuedTo}</Row>}
        </>
      )}
    </Panel>
  );
}
