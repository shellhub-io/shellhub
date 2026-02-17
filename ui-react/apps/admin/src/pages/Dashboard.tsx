import { useEffect } from "react";
import { useDevicesStore } from "../stores/devicesStore";
import { useSessionsStore } from "../stores/sessionsStore";
import { useNamespacesStore } from "../stores/namespacesStore";
import PageHeader from "../components/common/PageHeader";
import StatCard from "../components/common/StatCard";
import WelcomeScreen from "../components/common/WelcomeScreen";
import {
  ClockIcon,
  Squares2X2Icon,
  CheckCircleIcon,
  SignalIcon,
} from "@heroicons/react/24/outline";

export default function Dashboard() {
  const {
    totalCount: devicesCount,
    loading,
    fetch: fetchDevices,
  } = useDevicesStore();
  const { totalCount: sessionsCount, fetch: fetchSessions } =
    useSessionsStore();
  const { currentNamespace } = useNamespacesStore();

  useEffect(() => {
    fetchDevices(1, 1);
    fetchSessions(1, 1);
  }, [fetchDevices, fetchSessions]);

  // Namespace exists but 0 devices — show welcome/onboarding
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
        variant="decorated"
        icon={<Squares2X2Icon className="w-6 h-6" />}
        overline="Home"
        title={currentNamespace?.name ?? "Dashboard"}
        description="Manage your ShellHub namespace"
      />

      <div className="mb-6">
        <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-muted mb-4">
          Devices
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
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
            value="--"
            linkLabel="View online"
            linkTo="/devices"
            accent="text-accent-green"
          />
        </div>
        <div className="animate-slide-up" style={{ animationDelay: "160ms" }}>
          <StatCard
            icon={<ClockIcon className="w-7 h-7" />}
            title="Pending Devices"
            value={sessionsCount}
            linkLabel="View pending"
            linkTo="/devices"
            accent="text-accent-yellow"
          />
        </div>
      </div>
    </div>
  );
}
