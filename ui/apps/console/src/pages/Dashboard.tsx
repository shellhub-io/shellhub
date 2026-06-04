import { Link, useNavigate } from "react-router-dom";
import {
  ClockIcon,
  Squares2X2Icon,
  CheckCircleIcon,
  SignalIcon,
  CommandLineIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useSessions } from "@/hooks/useSessions";
import { useStats } from "@/hooks/useStats";
import { hasAnyDevices } from "@/utils/stats";
import { formatDate } from "@/utils/date";
import PageHeader from "@/components/common/PageHeader";
import StatCard from "@/components/common/StatCard";
import WelcomeScreen from "@/components/common/WelcomeScreen";
import CopyButton from "@/components/common/CopyButton";
import DeviceChip from "@/components/common/DeviceChip";
import DataTable, { type Column } from "@/components/common/DataTable";
import { sessionType } from "@/utils/session";
import type { Session } from "@/client";

export default function Dashboard() {
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const { sessions } = useSessions({ page: 1, perPage: 5 });
  const { stats, isLoading: statsLoading, error: statsError } = useStats();
  const navigate = useNavigate();

  if (statsLoading) return null;

  if (!statsError && stats && !hasAnyDevices(stats) && currentNamespace) {
    return <WelcomeScreen namespaceName={currentNamespace.name} />;
  }

  const goToPending = () => {
    void navigate("/devices?status=pending");
  };

  const sessionColumns: Column<Session>[] = [
    {
      key: "active",
      header: "Active",
      headerClassName: "w-14",
      render: (s) => (
        <span
          className={`w-2 h-2 rounded-full inline-block ${
            s.active
              ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
              : "bg-text-muted/40"
          }`}
        />
      ),
    },
    {
      key: "device",
      header: "Device",
      render: (s) =>
        s.device?.uid ? (
          <DeviceChip
            uid={s.device.uid}
            name={s.device.name ?? (s.device_uid ?? "").substring(0, 8)}
            online={s.device.online}
            osId={s.device.info?.id}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {s.device?.name ?? (s.device_uid ?? "").substring(0, 8)}
          </span>
        ),
    },
    {
      key: "username",
      header: "Username",
      render: (s) => {
        const suspicious = !s.authenticated;
        return (
          <div className="flex items-center gap-1.5">
            {suspicious && (
              <ExclamationTriangleIcon
                className="w-3.5 h-3.5 text-accent-red/70 shrink-0"
                strokeWidth={2}
                title="Not authenticated"
              />
            )}
            <code
              className={`text-xs font-mono ${suspicious ? "text-accent-red/60" : "text-text-secondary"}`}
            >
              {s.username}
            </code>
          </div>
        );
      },
    },
    {
      key: "type",
      header: "Type",
      render: (s) => {
        const type = sessionType(s);
        return type ? (
          <span
            className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
          >
            {type.label}
          </span>
        ) : (
          <span className="text-2xs text-text-muted">{"\u2014"}</span>
        );
      },
    },
    {
      key: "started",
      header: "Started",
      render: (s) => (
        <span className="text-xs text-text-secondary">
          {formatDate(s.started_at)}
        </span>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<Squares2X2Icon className="w-6 h-6" />}
        overline="Home"
        title={currentNamespace?.name ?? "Dashboard"}
        description="Manage your ShellHub namespace"
      >
        {currentNamespace && (
          <div className="flex items-center gap-2 bg-card border border-border rounded-lg px-3 py-2">
            <span className="text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted">
              Tenant ID
            </span>
            <span className="font-mono text-xs text-text-secondary">
              {currentNamespace.tenant_id}
            </span>
            <CopyButton text={currentNamespace.tenant_id} size="sm" />
          </div>
        )}
      </PageHeader>

      <div className="mb-4">
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-4">
          Devices
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-10">
        <div className="animate-slide-up" style={{ animationDelay: "0ms" }}>
          <StatCard
            icon={<CheckCircleIcon className="w-7 h-7" />}
            title="Accepted Devices"
            value={stats?.registered_devices ?? "--"}
            linkLabel="View all devices"
            linkTo="/devices"
          />
        </div>
        <div className="animate-slide-up" style={{ animationDelay: "80ms" }}>
          <StatCard
            icon={<SignalIcon className="w-7 h-7" />}
            title="Online Devices"
            value={stats?.online_devices ?? "--"}
            linkLabel="View all devices"
            linkTo="/devices"
            accent="text-accent-green"
          />
        </div>
        <div className="animate-slide-up" style={{ animationDelay: "160ms" }}>
          <StatCard
            icon={<ClockIcon className="w-7 h-7" />}
            title="Pending Devices"
            value={stats?.pending_devices ?? "--"}
            linkLabel="View pending"
            onClick={goToPending}
            accent="text-accent-yellow"
          />
        </div>
      </div>

      <div className="mb-4 flex items-center justify-between">
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Recent Sessions
        </p>
        <Link
          to="/sessions"
          className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
        >
          View all &rarr;
        </Link>
      </div>

      <div
        className="bg-card border border-border rounded-lg overflow-hidden animate-slide-up"
        style={{ animationDelay: "300ms" }}
      >
        <DataTable
          columns={sessionColumns}
          data={sessions}
          rowKey={(s) => s.uid}
          noWrapper
          onRowClick={(s) => void navigate(`/sessions/${s.uid}`)}
          // border-l-2 on every row (transparent by default) keeps the row
          // height stable when the red border appears on unauthenticated rows.
          rowClassName={(s) =>
            !s.authenticated
              ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
              : "border-l-2 border-l-transparent"
          }
          emptyState={
            <div className="flex flex-col items-center justify-center">
              <CommandLineIcon className="w-8 h-8 mb-3 opacity-40 text-text-muted" />
              <p className="text-sm text-text-muted">No recent sessions</p>
            </div>
          }
        />
      </div>
    </div>
  );
}
