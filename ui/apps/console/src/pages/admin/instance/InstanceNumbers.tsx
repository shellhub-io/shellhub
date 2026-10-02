import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Card } from "@shellhub/design-system/primitives";
import { useAdminStats } from "@/hooks/useAdminStats";
import { useAdminNamespaces } from "@/hooks/useAdminNamespaces";
import { useAdminUsers } from "@/hooks/useAdminUsers";
import { LABEL_BASE } from "@/utils/styles";
import { formatCount } from "@/utils/count";
import { Count } from "./Panel";

function Metric({
  label,
  value,
  isLoading,
  isError,
  detail,
  to,
}: {
  label: string;
  value?: number;
  isLoading: boolean;
  isError: boolean;
  detail?: ReactNode;
  to?: string;
}) {
  const body = (
    <>
      <p className={LABEL_BASE}>{label}</p>
      <p className="mt-3 text-3xl font-bold text-text-primary">
        <Count value={value} isLoading={isLoading} isError={isError} />
      </p>
      <p className="mt-1 text-xs text-text-muted min-h-4">{detail}</p>
    </>
  );

  return to ? (
    <Card
      as={Link}
      to={to}
      hover
      className="rounded-lg p-5 block hover:border-primary/30"
    >
      {body}
    </Card>
  ) : (
    <Card className="rounded-lg p-5">{body}</Card>
  );
}

function percent(part?: number, whole?: number) {
  if (part === undefined || !whole) return undefined;
  return Math.round((part / whole) * 100);
}

/**
 * The instance's size and live load in figures, for an instance too large to read person by
 * person. A figure that could not be loaded shows a dash and the strip says so.
 */
export default function InstanceNumbers() {
  const {
    stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();
  const namespaces = useAdminNamespaces({ perPage: 1 });
  const unconfirmed = useAdminUsers({ perPage: 1, subset: "not_confirmed" });

  const namespacesKnown = !namespaces.isLoading && !namespaces.isError;
  const onlineShare = percent(stats?.online_devices, stats?.registered_devices);
  const devicesPerNamespace =
    stats?.registered_devices !== undefined &&
    namespacesKnown &&
    namespaces.totalCount > 0
      ? (stats.registered_devices / namespaces.totalCount).toFixed(1)
      : undefined;
  const unconfirmedKnown = !unconfirmed.isLoading && !unconfirmed.isError;

  return (
    <section aria-label="Instance numbers">
      {(statsError || namespaces.isError || unconfirmed.isError) && (
        <p role="alert" className="mb-3 text-xs text-accent-red">
          Couldn&apos;t load some of these numbers.
        </p>
      )}
      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-4">
        <Metric
          label="Users"
          value={stats?.registered_users}
          isLoading={statsLoading}
          isError={statsError}
          to="/admin/users"
          detail={
            unconfirmedKnown &&
            `${formatCount(unconfirmed.totalCount)} haven't confirmed their email`
          }
        />
        <Metric
          label="Namespaces"
          value={namespaces.totalCount}
          isLoading={namespaces.isLoading}
          isError={namespaces.isError}
          to="/admin/namespaces"
          detail={
            devicesPerNamespace && `${devicesPerNamespace} devices on average`
          }
        />
        <Metric
          label="Devices"
          value={stats?.registered_devices}
          isLoading={statsLoading}
          isError={statsError}
          detail={
            stats?.pending_devices !== undefined &&
            stats.rejected_devices !== undefined &&
            `${formatCount(stats.pending_devices)} pending · ${formatCount(stats.rejected_devices)} rejected`
          }
        />
        <Metric
          label="Online now"
          value={stats?.online_devices}
          isLoading={statsLoading}
          isError={statsError}
          detail={
            onlineShare !== undefined && `${onlineShare}% of accepted devices`
          }
        />
        <Metric
          label="Active sessions"
          value={stats?.active_sessions}
          isLoading={statsLoading}
          isError={statsError}
          detail="Open right now"
        />
      </div>
    </section>
  );
}
