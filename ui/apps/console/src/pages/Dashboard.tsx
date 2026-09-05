import {
  ClockIcon,
  Squares2X2Icon,
  CheckCircleIcon,
  SignalIcon,
} from "@heroicons/react/24/outline";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useStats } from "@/hooks/useStats";
import { hasAnyDevices } from "@/utils/stats";
import PageHeader from "@/components/common/PageHeader";
import StatCard from "@/components/common/StatCard";
import WelcomeScreen from "@/components/common/WelcomeScreen";
import CopyButton from "@/components/common/CopyButton";
import RecentSessionsTable from "@/components/sessions/RecentSessionsTable";
import { Card } from "@shellhub/design-system/primitives";

/**
 * The namespace dashboard: device counts, recent sessions, and the first-run wizard for a
 * namespace with nothing in it yet.
 */
export default function Dashboard() {
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const { stats, isLoading: statsLoading, error: statsError } = useStats();

  if (statsLoading) return null;

  if (!statsError && stats && !hasAnyDevices(stats) && currentNamespace) {
    return <WelcomeScreen namespaceName={currentNamespace.name} />;
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
          <Card className="flex items-center gap-2 px-3 py-2">
            <span className="text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted">
              Tenant ID
            </span>
            <span className="font-mono text-xs text-text-secondary">
              {currentNamespace.tenant_id}
            </span>
            <CopyButton text={currentNamespace.tenant_id} size="sm" />
          </Card>
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
            linkLabel="Review pending"
            linkTo="/pending-devices"
            accent="text-accent-yellow"
          />
        </div>
      </div>

      <RecentSessionsTable />
    </div>
  );
}
