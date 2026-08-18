import type { ReactNode } from "react";
import {
  UsersIcon,
  CpuChipIcon,
  SignalIcon,
  ClockIcon,
  XCircleIcon,
  CommandLineIcon,
  ChartBarIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import PageHeader from "@/components/common/PageHeader";
import StatCard from "@/components/common/StatCard";
import RecentSessionsTable from "@/components/sessions/RecentSessionsTable";
import { useAdminStats } from "@/hooks/useAdminStats";
import PageLoader from "@/components/common/PageLoader";

export default function AdminDashboard() {
  const {
    stats: statsData,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();

  if (statsLoading) {
    return <PageLoader label="Loading dashboard statistics" padding="fill" />;
  }

  if (statsError) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center" role="alert">
          <ExclamationCircleIcon className="w-10 h-10 text-accent-red mx-auto mb-3" />
          <p className="text-sm font-medium text-text-primary">
            Failed to load dashboard statistics
          </p>
          <p className="text-2xs text-text-muted mt-1">
            Please try again later.
          </p>
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

      <RecentSessionsTable isAdmin />
    </div>
  );
}
