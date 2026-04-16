import { useParams, Link } from "react-router-dom";
import {
  ChevronRightIcon,
  CpuChipIcon,
  InformationCircleIcon,
  ComputerDesktopIcon,
  ClockIcon,
  TagIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import { useAdminDevice } from "@/hooks/useAdminDevices";
import CopyButton from "@/components/common/CopyButton";
import DistroIcon from "@/components/common/DistroIcon";
import PlatformBadge from "@/components/common/PlatformBadge";
import DeviceStatusChip from "./DeviceStatusChip";
import { formatDateFull, formatRelative } from "@/utils/date";

const LABEL
  = "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";

export default function AdminDeviceDetails() {
  const { uid } = useParams<{ uid: string }>();
  const { data: device, isLoading, error } = useAdminDevice(uid ?? "");

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24" role="status">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading device details</span>
      </div>
    );
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
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/devices"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Devices
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary">
          {device.name}
        </span>
      </nav>

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div className="relative w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
          <CpuChipIcon className="w-7 h-7 text-primary" />
          {/* Online indicator dot */}
          <span
            className={`absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full border-2 border-surface ${
              device.online ? "bg-accent-green" : "bg-text-muted/30"
            }`}
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
              className={`inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md ${
                device.online
                  ? "bg-accent-green/10 text-accent-green border border-accent-green/20"
                  : "bg-text-muted/10 text-text-muted border border-text-muted/20"
              }`}
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
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Identity
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>UID</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <span
                  className="text-xs font-mono text-text-primary truncate max-w-[180px]"
                  title={device.uid}
                >
                  {device.uid}
                </span>
                <CopyButton text={device.uid} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>MAC Address</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <span className="text-xs font-mono text-text-primary">
                  {device.identity?.mac ?? "\u2014"}
                </span>
                {device.identity?.mac && (
                  <CopyButton text={device.identity.mac} />
                )}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Remote Address</dt>
              <dd className={VALUE}>
                <span className="font-mono text-xs">
                  {device.remote_addr ?? "\u2014"}
                </span>
              </dd>
            </div>
          </dl>
        </div>

        {/* System Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ComputerDesktopIcon className="w-4 h-4 text-primary" />
            System
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Operating System</dt>
              <dd className="flex items-center gap-2 mt-0.5">
                <DistroIcon
                  id={device.info?.id ?? ""}
                  className="text-base leading-none"
                />
                <span className="text-sm text-text-primary">
                  {device.info?.pretty_name ?? "\u2014"}
                </span>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Architecture</dt>
              <dd className={VALUE}>
                <span className="font-mono text-xs">
                  {device.info?.arch ?? "\u2014"}
                </span>
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Platform</dt>
              <dd className="mt-0.5">
                {device.info?.platform ? (
                  <PlatformBadge platform={device.info.platform} />
                ) : (
                  <span className="text-sm text-text-muted">&mdash;</span>
                )}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Agent Version</dt>
              <dd className={VALUE}>
                <span className="font-mono text-xs">
                  {device.info?.version ?? "\u2014"}
                </span>
              </dd>
            </div>
          </dl>
        </div>

        {/* Namespace & Timeline Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <ClockIcon className="w-4 h-4 text-primary" />
            Namespace & Timeline
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Namespace</dt>
              <dd className="mt-0.5">
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
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Tenant ID</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <span className="text-xs font-mono text-text-primary">
                  {device.tenant_id}
                </span>
                <CopyButton text={device.tenant_id} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Created</dt>
              <dd className={VALUE}>{formatDateFull(device.created_at)}</dd>
            </div>
            <div>
              <dt className={LABEL}>Last Seen</dt>
              <dd className={VALUE}>{formatRelative(device.last_seen)}</dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Tags Section */}
      <div className="bg-card border border-border rounded-xl p-5 mb-6">
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
      </div>

      {/* Public Key Section */}
      {device.public_key && (
        <div className="bg-card border border-border rounded-xl p-5">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2 mb-3">
            <KeyIcon className="w-4 h-4 text-primary" />
            Public Key
          </h3>
          <pre className="text-2xs font-mono text-text-secondary bg-surface border border-border rounded-lg p-4 overflow-x-auto whitespace-pre-wrap break-all">
            {device.public_key}
          </pre>
        </div>
      )}
    </div>
  );
}
