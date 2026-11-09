import { useState, type CSSProperties } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { ArrowLeftIcon, MapIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useQuery } from "@tanstack/react-query";
import { Spinner } from "@shellhub/design-system/primitives";
import { getDevices, getDevicesQueryKey } from "@/client";
import { useDeviceLocations } from "@/hooks/useDeviceRequests";
import { paginatedQueryOptions } from "@/api/pagination";
import AreaDevicesPanel from "./AreaDevicesPanel";
import LocationsMap from "./LocationsMap";
import {
  areaWithin,
  compactCount,
  formatBounds,
  modeTotal,
  ONLINE_COLOR_STOPS,
  parseBounds,
  type MapMode,
} from "./mapModel";

const [[OFFLINE_SHARE], [PARTIAL_SHARE], [ONLINE_SHARE]] = ONLINE_COLOR_STOPS;
const PARTIAL_AT =
  ((PARTIAL_SHARE - OFFLINE_SHARE) / (ONLINE_SHARE - OFFLINE_SHARE)) * 100;

function percent(share: number): string {
  return `${Math.round(share * 100)}%`;
}

const MODES: { value: MapMode; label: string }[] = [
  { value: "all", label: "All" },
  { value: "online", label: "Online" },
  { value: "offline", label: "Offline" },
];

function Stat({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent?: string;
}) {
  return (
    <div>
      <p className="text-3xs font-mono font-semibold uppercase tracking-label text-text-muted">
        {label}
      </p>
      <p
        className={cn(
          "mt-0.5 font-mono text-lg font-semibold tabular-nums text-text-primary",
          accent,
        )}
      >
        {value}
      </p>
    </div>
  );
}

/**
 * The device map: every accepted device with a known position, clustered on a world map and
 * refreshed every thirty seconds. Picking a cluster or point lists the devices in it, or opens the
 * brief of a point's one device. The area picked lives in the URL as ?area=west,south,east,north,
 * so a link or a reload opens the map on it once a read of the locations has succeeded; a value
 * that does not parse is ignored. The devices with no position are counted against the device
 * list, which leaves connector containers out as the map does.
 */
export default function DeviceMap() {
  const [mode, setMode] = useState<MapMode>("all");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const { locations, source, isLoading, read, error } =
    useDeviceLocations();
  const picked = parseBounds(searchParams.get("area"));
  const area = picked && read ? areaWithin(locations, picked) : null;
  const { data: listed } = useQuery(
    paginatedQueryOptions(getDevices, getDevicesQueryKey, {
      query: { status: "accepted", page: 1, per_page: 1 },
    }),
  );

  const located = modeTotal(locations, "all");
  const online = modeTotal(locations, "online");
  const unlocated = Math.max(0, (listed?.totalCount ?? located) - located);

  return (
    <div className="absolute inset-0 overflow-hidden bg-background">
      <LocationsMap
        locations={locations}
        source={source}
        mode={mode}
        selected={area?.bounds}
        onSelect={(selected) => {
          void navigate(
            { search: `?area=${formatBounds(selected.bounds)}` },
            { replace: true },
          );
        }}
        navigable
      />

      <section className="absolute left-4 top-4 z-10 w-72 rounded-xl border border-border bg-surface/95 p-4 shadow-float backdrop-blur">
        <Link
          to="/dashboard"
          className="mb-3 inline-flex items-center gap-1 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted transition-colors hover:text-primary"
        >
          <ArrowLeftIcon className="h-3 w-3" />
          Dashboard
        </Link>
        <div className="flex items-center gap-2">
          <MapIcon className="h-5 w-5 text-primary" />
          <h1 className="text-sm font-semibold text-text-primary">
            Device Map
          </h1>
          {isLoading && <Spinner size="sm" className="ml-auto" />}
        </div>
        <p className="mt-1 text-2xs leading-relaxed text-text-muted">
          Positions come from a GeoIP lookup of the address each device connects
          from, so they are accurate to the city at best.
        </p>

        <div className="mt-4 grid grid-cols-3 gap-3">
          <Stat label="Located" value={compactCount(located)} />
          <Stat
            label="Online"
            value={compactCount(online)}
            accent="text-accent-green"
          />
          <Stat
            label="No position"
            value={compactCount(unlocated)}
            accent="text-text-secondary"
          />
        </div>

        <div
          role="radiogroup"
          aria-label="Devices shown"
          className="mt-4 grid grid-cols-3 rounded-lg border border-border bg-background/60 p-0.5"
        >
          {MODES.map((option) => (
            <button
              key={option.value}
              type="button"
              role="radio"
              aria-checked={mode === option.value}
              onClick={() => setMode(option.value)}
              className={cn(
                "rounded-md py-1 text-xs font-medium text-text-secondary transition-colors hover:text-text-primary",
                mode === option.value && "bg-card text-text-primary shadow-sm",
              )}
            >
              {option.label}
            </button>
          ))}
        </div>

        {mode === "all" && (
          <div className="mt-4">
            <div
              className="h-1.5 rounded-full bg-gradient-to-r from-accent-red via-accent-yellow to-accent-green"
              style={
                {
                  "--tw-gradient-via-position": `${PARTIAL_AT}%`,
                } as CSSProperties
              }
            />
            <div className="relative mt-1 h-3 font-mono text-3xs text-text-muted">
              <span className="absolute left-0">
                ≤{percent(OFFLINE_SHARE)} online
              </span>
              <span
                className="absolute -translate-x-1/2"
                style={{ left: `${PARTIAL_AT}%` }}
              >
                {percent(PARTIAL_SHARE)}
              </span>
              <span className="absolute right-0">{percent(ONLINE_SHARE)}</span>
            </div>
          </div>
        )}

        {error && (
          <p className="mt-3 text-xs text-accent-red">
            Could not load device locations.
          </p>
        )}
        {!isLoading && !error && locations.length === 0 && (
          <p className="mt-3 text-xs text-text-secondary">
            No accepted device has a known position yet.
          </p>
        )}
      </section>

      {area && (
        <AreaDevicesPanel
          key={area.bounds.join(",")}
          area={area}
          onClose={() => {
            void navigate({ search: "" }, { replace: true });
          }}
        />
      )}
    </div>
  );
}
