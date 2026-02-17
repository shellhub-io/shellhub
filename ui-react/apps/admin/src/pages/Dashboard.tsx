import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ClockIcon,
  Squares2X2Icon,
  CheckCircleIcon,
  SignalIcon,
  CommandLineIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import { useDevicesStore } from "../stores/devicesStore";
import { useNamespacesStore } from "../stores/namespacesStore";
import { useSessionsStore } from "../stores/sessionsStore";
import { getStats } from "../api/stats";
import { type Stats } from "../types/stats";
import { formatDate } from "../utils/date";
import PageHeader from "../components/common/PageHeader";
import StatCard from "../components/common/StatCard";
import WelcomeScreen from "../components/common/WelcomeScreen";
import CopyButton from "../components/common/CopyButton";
import DeviceChip from "../components/common/DeviceChip";
import { sessionType } from "../utils/session";
import { TH } from "../utils/styles";

export default function Dashboard() {
  const {
    totalCount: devicesCount,
    loading,
    fetch: fetchDevices,
  } = useDevicesStore();
  const { currentNamespace } = useNamespacesStore();
  const { sessions, fetch: fetchSessions } = useSessionsStore();
  const [stats, setStats] = useState<Stats | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDevices(1, 1);
    fetchSessions(1, 5);
    getStats()
      .then(setStats)
      .catch(() => {});
  }, [fetchDevices, fetchSessions]);

  if (!loading && devicesCount === 0 && currentNamespace) {
    return (
      <WelcomeScreen
        namespaceName={currentNamespace.name}
        tenantId={currentNamespace.tenant_id}
      />
    );
  }

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
            value={devicesCount}
            linkLabel="View all devices"
            linkTo="/devices"
          />
        </div>
        <div className="animate-slide-up" style={{ animationDelay: "80ms" }}>
          <StatCard
            icon={<SignalIcon className="w-7 h-7" />}
            title="Online Devices"
            value={stats?.online_devices ?? "--"}
            linkLabel="View online"
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
            linkTo="/devices"
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
        {sessions.length > 0 ? (
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-surface/50">
                <th className={`${TH} w-14`}>Active</th>
                <th className={TH}>Device</th>
                <th className={TH}>Username</th>
                <th className={TH}>Type</th>
                <th className={TH}>Started</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {sessions.map((session) => {
                const type = sessionType(session);
                const suspicious = !session.authenticated;
                return (
                  <tr
                    key={session.uid}
                    onClick={() => navigate(`/sessions/${session.uid}`)}
                    className={`transition-colors cursor-pointer ${
                      suspicious
                        ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
                        : "hover:bg-hover-subtle border-l-2 border-l-transparent"
                    }`}
                  >
                    <td className="px-4 py-3.5">
                      <span
                        className={`w-2 h-2 rounded-full inline-block ${
                          session.active
                            ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
                            : "bg-text-muted/40"
                        }`}
                      />
                    </td>
                    <td className="px-4 py-3.5">
                      {session.device?.uid ? (
                        <DeviceChip
                          uid={session.device.uid}
                          name={
                            session.device.name ??
                            session.device_uid.substring(0, 8)
                          }
                          online={session.device.online}
                          osId={session.device.info?.id}
                          onClick={(e) => e.stopPropagation()}
                        />
                      ) : (
                        <span className="text-xs font-mono text-text-primary">
                          {session.device?.name ??
                            session.device_uid.substring(0, 8)}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3.5">
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
                          {session.username}
                        </code>
                      </div>
                    </td>
                    <td className="px-4 py-3.5">
                      {type ? (
                        <span
                          className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
                        >
                          {type.label}
                        </span>
                      ) : (
                        <span className="text-2xs text-text-muted">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {formatDate(session.started_at)}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <CommandLineIcon className="w-8 h-8 mb-3 opacity-40" />
            <p className="text-sm">No recent sessions</p>
          </div>
        )}
      </div>
    </div>
  );
}
