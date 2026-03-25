import type { ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  UsersIcon,
  CpuChipIcon,
  SignalIcon,
  ClockIcon,
  XCircleIcon,
  CommandLineIcon,
  ChartBarIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import PageHeader from "../../components/common/PageHeader";
import StatCard from "../../components/common/StatCard";
import DeviceChip from "../../components/common/DeviceChip";
import { useAdminStats } from "../../hooks/useAdminStats";
import { useAdminSessions } from "../../hooks/useAdminSessions";
import { formatDate } from "../../utils/date";
import { sessionType } from "../../utils/session";
import { TH } from "../../utils/styles";

export default function AdminDashboard() {
  const { stats: statsData, isLoading: statsLoading, isError: statsError } = useAdminStats();
  const { sessions, isLoading: sessionsLoading, error: sessionsError } = useAdminSessions();
  const navigate = useNavigate();

  if (statsLoading) {
    return (
      <div
        className="flex-1 flex items-center justify-center"
        role="status"
        aria-label="Loading dashboard statistics"
      >
        <span
          className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin"
          aria-hidden="true"
        />
      </div>
    );
  }

  if (statsError) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center" role="alert">
          <ExclamationCircleIcon className="w-10 h-10 text-accent-red mx-auto mb-3" />
          <p className="text-sm font-medium text-text-primary">
            Failed to load dashboard statistics
          </p>
          <p className="text-2xs text-text-muted mt-1">Please try again later.</p>
        </div>
      </div>
    );
  }

  const stats = statsData ?? {};

  const statCards: Array<{
    value: number;
    icon: ReactNode;
    title: string;
    linkLabel: string;
    linkTo: string;
    accent?: string;
  }> = [
    {
      value: stats.registered_users ?? 0,
      icon: <UsersIcon className="w-7 h-7" />,
      title: "Registered Users",
      linkLabel: "View all Users",
      linkTo: "/admin/users",
    },
    {
      value: stats.registered_devices ?? 0,
      icon: <CpuChipIcon className="w-7 h-7" />,
      title: "Registered Devices",
      linkLabel: "View all Devices",
      linkTo: "/admin/devices",
    },
    {
      value: stats.online_devices ?? 0,
      icon: <SignalIcon className="w-7 h-7" />,
      title: "Online Devices",
      linkLabel: "View Online Devices",
      linkTo: "/admin/devices",
      accent: "text-accent-green",
    },
    {
      value: stats.pending_devices ?? 0,
      icon: <ClockIcon className="w-7 h-7" />,
      title: "Pending Devices",
      linkLabel: "View Pending Devices",
      linkTo: "/admin/devices",
      accent: "text-accent-yellow",
    },
    {
      value: stats.rejected_devices ?? 0,
      icon: <XCircleIcon className="w-7 h-7" />,
      title: "Rejected Devices",
      linkLabel: "View Rejected Devices",
      linkTo: "/admin/devices",
      accent: "text-accent-red",
    },
    {
      value: stats.active_sessions ?? 0,
      icon: <CommandLineIcon className="w-7 h-7" />,
      title: "Active Sessions",
      linkLabel: "View all Sessions",
      linkTo: "/admin/sessions",
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ChartBarIcon className="w-6 h-6" />}
        overline="Admin Dashboard"
        title="System Overview"
        description="Monitor key metrics about users, devices, and sessions across the instance."
      />

      <div className="mb-4">
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-4">
          Stats
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-10">
        {statCards.map((card, i) => (
          <div
            key={card.title}
            className="animate-slide-up"
            style={{ animationDelay: `${i * 80}ms` }}
          >
            <StatCard
              icon={card.icon}
              title={card.title}
              value={card.value}
              linkLabel={card.linkLabel}
              linkTo={card.linkTo}
              accent={card.accent}
            />
          </div>
        ))}
      </div>

      {!sessionsLoading && !sessionsError && (
        <>
          <div className="mb-4 flex items-center justify-between">
            <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
              Recent Sessions
            </p>
            <Link
              to="/admin/sessions"
              className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
            >
              View all &rarr;
            </Link>
          </div>

          <div
            className="bg-card border border-border rounded-lg overflow-hidden animate-slide-up"
            style={{ animationDelay: "560ms" }}
          >
            {sessions.length > 0
              ? (
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
                          onClick={() => void navigate(`/admin/sessions/${session.uid}`)}
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
                            {session.device
                              ? (
                                <DeviceChip
                                  disableLink
                                  name={session.device.name}
                                  online={session.device.online}
                                  osId={session.device.info?.id}
                                />
                              )
                              : (
                                <span className="text-xs font-mono text-text-primary">
                                  {(session.device_uid ?? "").substring(0, 8)}
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
                                className={`text-xs font-mono ${
                                  suspicious ? "text-accent-red/60" : "text-text-secondary"
                                }`}
                              >
                                {session.username}
                              </code>
                            </div>
                          </td>
                          <td className="px-4 py-3.5">
                            {type
                              ? (
                                <span
                                  className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
                                >
                                  {type.label}
                                </span>
                              )
                              : (
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
              )
              : (
                <div className="flex flex-col items-center justify-center py-12 text-text-muted">
                  <CommandLineIcon className="w-8 h-8 mb-3 opacity-40" />
                  <p className="text-sm">No recent sessions</p>
                </div>
              )}
          </div>
        </>
      )}
    </div>
  );
}
