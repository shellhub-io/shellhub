import { useParams, Link } from "react-router-dom";
import {
  CpuChipIcon,
  InformationCircleIcon,
  ComputerDesktopIcon,
  ClockIcon,
  TagIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useAdminDevice } from "@/hooks/useAdminDevices";
import Breadcrumb from "@/components/common/Breadcrumb";
import DistroIcon from "@/components/common/DistroIcon";
import PlatformBadge from "@/components/common/PlatformBadge";
import DeviceStatusChip from "./DeviceStatusChip";
import { formatDateFull, formatRelative } from "@/utils/date";
import InfoItem from "@/components/common/InfoItem";
import PageLoader from "@/components/common/PageLoader";
import { Card } from "@shellhub/design-system/primitives";

export default function AdminDeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const { data: device, isLoading, error } = useAdminDevice(uid ?? "");

  if (isLoading) {
    return <PageLoader label="Loading device details" />;
  }

  if (error || !device) {
    return (
      <div className="text-center py-24">
        <CpuChipIcon
          className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
          strokeWidth={1}
        />
        <p className="text-sm text-text-muted mb-2">Device not found</p>
        <Link
          to="/admin/devices"
          className="text-sm text-primary hover:underline"
        >
          Back to devices
        </Link>
      </div>
    );
  }

  const tags = device.tags ?? [];

  return (
    <div className="animate-fade-in">
      <Breadcrumb
        items={[
          { label: "Devices", to: "/admin/devices" },
          { label: device.name },
        ]}
      />

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div className="relative w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
          <CpuChipIcon className="w-7 h-7 text-primary" />
          {/* Online indicator dot */}
          <span
            className={cn(
              "absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full border-2 border-surface",
              device.online ? "bg-accent-green" : "bg-text-muted/30",
            )}
            title={device.online ? "Online" : "Offline"}
            aria-label={device.online ? "Online" : "Offline"}
          />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-text-primary">
            {device.name}
          </h1>
          <div className="flex items-center gap-2 mt-1.5">
            <span
              className={cn(
                "inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md",
                device.online
                  ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                  : "bg-text-muted/10 text-text-muted border border-text-muted/20",
              )}
            >
              {device.online ? "Online" : "Offline"}
            </span>
            <DeviceStatusChip status={device.status} />
          </div>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {/* Identity Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Identity
          </h3>
          <dl className="space-y-3">
            <InfoItem
              label="UID"
              value={device.uid}
              mono
              copyable
              truncate={8}
            />
            <InfoItem
              label="MAC Address"
              value={device.identity?.mac ?? ""}
              mono
              copyable
            />
            <InfoItem
              label="Remote Address"
              value={device.remote_addr ?? ""}
              mono
            />
          </dl>
        </Card>

        {/* System Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ComputerDesktopIcon className="w-4 h-4 text-primary" />
            System
          </h3>
          <dl className="space-y-3">
            <InfoItem label="Operating System">
              <DistroIcon
                id={device.info?.id ?? ""}
                className="text-base leading-none"
              />
              <span className="text-sm text-text-primary">
                {device.info?.pretty_name ?? "\u2014"}
              </span>
            </InfoItem>
            <InfoItem
              label="Architecture"
              value={device.info?.arch ?? ""}
              mono
            />
            <InfoItem label="Platform">
              {device.info?.platform ? (
                <PlatformBadge platform={device.info.platform} />
              ) : (
                <span className="text-sm text-text-muted">&mdash;</span>
              )}
            </InfoItem>
            <InfoItem
              label="Agent Version"
              value={device.info?.version ?? ""}
              mono
            />
          </dl>
        </Card>

        {/* Namespace & Timeline Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Namespace & Timeline
          </h3>
          <dl className="space-y-3">
            <InfoItem label="Namespace">
              {device.namespace ? (
                <Link
                  to={`/admin/namespaces/${device.tenant_id}`}
                  className="text-sm text-primary hover:underline"
                >
                  {device.namespace}
                </Link>
              ) : (
                <span className="text-sm text-text-muted">&mdash;</span>
              )}
            </InfoItem>
            <InfoItem
              label="Tenant ID"
              value={device.tenant_id}
              mono
              copyable
            />
            <InfoItem
              label="Created"
              value={formatDateFull(device.created_at)}
            />
            <InfoItem
              label="Last Seen"
              value={formatRelative(device.last_seen)}
            />
          </dl>
        </Card>
      </div>

      {/* Tags Section */}
      <Card className="p-5 mb-6">
        <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2 mb-3">
          <TagIcon className="w-4 h-4 text-primary" />
          Tags
        </h3>
        {tags.length > 0 ? (
          <div className="flex items-center gap-1.5 flex-wrap">
            {tags.map((tag, i) => (
              <span
                key={`${tag}-${i}`}
                className="inline-flex items-center px-2 py-0.5 bg-primary/8 text-primary text-2xs rounded-md font-medium border border-primary/15"
              >
                {tag}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs font-mono text-text-muted/50">No tags</p>
        )}
      </Card>

      {/* Public Key Section */}
      {device.public_key && (
        <Card className="p-5">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2 mb-3">
            <KeyIcon className="w-4 h-4 text-primary" />
            Public Key
          </h3>
          <pre className="text-2xs font-mono text-text-secondary bg-surface border border-border rounded-lg p-4 overflow-x-auto whitespace-pre-wrap break-all">
            {device.public_key}
          </pre>
        </Card>
      )}
    </div>
  );
}
