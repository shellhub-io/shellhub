import { Link, useNavigate } from "react-router-dom";
import { Squares2X2Icon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Card } from "@shellhub/design-system/primitives";
import { useNamespace, useStats } from "@/hooks/useNamespaceRequests";
import { useAuthStore } from "@/stores/authStore";
import { useDeviceLocations } from "@/hooks/useDeviceRequests";
import PageHeader from "@/components/common/PageHeader";
import CopyButton from "@/components/common/CopyButton";
import RecentSessionsTable from "@/components/sessions/RecentSessionsTable";
import { formatCount } from "@/utils/count";
import LocationsMap from "./device-map/LocationsMap";
import { formatBounds } from "./device-map/mapModel";

interface Figure {
  label: string;
  value: string;
  detail?: string;
  accent?: string;
  to?: string;
}

function FigureCell({ label, value, detail, accent, to }: Figure) {
  const body = (
    <>
      <span className="block text-3xs font-mono font-semibold uppercase tracking-label text-text-muted">
        {label}
      </span>
      <span className="mt-1 flex items-baseline gap-2">
        <span
          className={cn(
            "font-mono text-2xl font-semibold tabular-nums",
            accent ?? "text-text-primary",
          )}
        >
          {value}
        </span>
        {detail && (
          <span className="font-mono text-xs text-text-muted">{detail}</span>
        )}
      </span>
    </>
  );

  if (!to) return <div className="px-5 py-3">{body}</div>;

  return (
    <Link
      to={to}
      className="block px-5 py-3 transition-colors hover:bg-hover-medium"
    >
      {body}
    </Link>
  );
}

function FigureStrip({
  figures,
  mapAvailable,
}: {
  figures: Figure[];
  mapAvailable: boolean;
}) {
  return (
    <div className="absolute left-4 top-4 z-10 flex divide-x divide-border overflow-hidden rounded-xl border border-border bg-surface/95 shadow-float backdrop-blur">
      {figures.map((figure) => (
        <FigureCell key={figure.label} {...figure} />
      ))}
      {mapAvailable ? (
        <Link
          to="/dashboard/map"
          className="flex items-center px-5 text-xs font-medium text-primary transition-colors hover:bg-hover-medium"
        >
          View full map &rarr;
        </Link>
      ) : (
        <span
          aria-disabled="true"
          className="flex cursor-not-allowed items-center px-5 text-xs font-medium text-text-muted opacity-60"
        >
          View full map &rarr;
        </span>
      )}
    </div>
  );
}

/**
 * The namespace dashboard: the device counts laid over a map of where the devices are, and the
 * recent sessions. Picking a point or cluster on the map opens the full map on it.
 */
export default function Dashboard() {
  const navigate = useNavigate();
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const { stats, isLoading: statsLoading } = useStats();
  const {
    locations,
    source,
    isLoading: locationsLoading,
    error: locationsError,
  } = useDeviceLocations();

  if (statsLoading) return null;

  const accepted = stats?.registered_devices ?? 0;
  const online = stats?.online_devices ?? 0;
  const offline = Math.max(0, accepted - online);
  const share = (count: number) =>
    stats && accepted > 0
      ? `${Math.round((count / accepted) * 100)}%`
      : undefined;
  const mapAvailable = locations.length > 0 && !locationsError;
  const mapNotice = locationsError
    ? "Could not load device positions."
    : !locationsLoading && !mapAvailable
      ? "No device has a position yet. Positions come from the address each device connects from."
      : undefined;
  const figures: Figure[] = [
    {
      label: "Online",
      value: stats ? formatCount(online) : "--",
      detail: share(online),
      accent: "text-accent-green",
      to: "/devices",
    },
    {
      label: "Offline",
      value: stats ? formatCount(offline) : "--",
      detail: share(offline),
      accent: "text-text-secondary",
      to: "/devices",
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<Squares2X2Icon className="w-6 h-6" />}
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

      <section className="relative mb-10 h-[26rem] overflow-hidden rounded-xl border border-border bg-card animate-slide-up">
        <LocationsMap
          locations={locations}
          source={source}
          mode="all"
          onSelect={(area) => {
            if (!area) return;
            void navigate(`/dashboard/map?area=${formatBounds(area.bounds)}`);
          }}
        />
        <FigureStrip figures={figures} mapAvailable={mapAvailable} />
        {mapNotice && (
          <p className="pointer-events-none absolute bottom-4 left-1/2 z-10 -translate-x-1/2 rounded-full border border-border bg-surface/95 px-4 py-1.5 text-xs text-text-secondary shadow-float backdrop-blur">
            {mapNotice}
          </p>
        )}
      </section>

      <RecentSessionsTable />
    </div>
  );
}
