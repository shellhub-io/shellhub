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
import StatCard, { type StatCardProps } from "@/components/common/StatCard";
import { useAdminStats } from "@/hooks/useAdminStats";
import PageLoader from "@/components/common/PageLoader";

/**
 * The admin dashboard: instance-wide counts of users, devices and sessions.
 */
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

  const statCards: StatCardProps[] = [
    {
      value: stats.registered_users ?? 0,
      icon: <UsersIcon className="w-7 h-7" />,
      title: "Registered Users",
      action: { label: "View all Users", to: "/admin/users" },
    },
    {
      value: stats.registered_devices ?? 0,
      icon: <CpuChipIcon className="w-7 h-7" />,
      title: "Registered Devices",
    },
    {
      value: stats.online_devices ?? 0,
      icon: <SignalIcon className="w-7 h-7" />,
      title: "Online Devices",
      accent: "text-accent-green",
    },
    {
      value: stats.pending_devices ?? 0,
      icon: <ClockIcon className="w-7 h-7" />,
      title: "Pending Devices",
      accent: "text-accent-yellow",
    },
    {
      value: stats.rejected_devices ?? 0,
      icon: <XCircleIcon className="w-7 h-7" />,
      title: "Rejected Devices",
      accent: "text-accent-red",
    },
    {
      value: stats.active_sessions ?? 0,
      icon: <CommandLineIcon className="w-7 h-7" />,
      title: "Active Sessions",
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ChartBarIcon className="w-6 h-6" />}
        title="System Overview"
        description="Monitor key metrics about users, devices, and sessions across the instance."
      />

      <div className="mb-4">
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-4">
          Stats
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {statCards.map((card, i) => (
          <div
            key={card.title}
            className="animate-slide-up"
            style={{ animationDelay: `${i * 80}ms` }}
          >
            <StatCard {...card} />
          </div>
        ))}
      </div>
    </div>
  );
}
