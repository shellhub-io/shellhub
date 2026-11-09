import { useEffect, useId, useRef, useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { ChevronUpIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import {
  IconButton,
  Spinner,
  StatusDot,
} from "@shellhub/design-system/primitives";
import { getDevices, getDevicesQueryKey, type Device } from "@/client";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { useWindowShown } from "@/stores/terminalStore";
import { useLoadOnReach } from "@/hooks/useLoadOnReach";
import { toBase64Json } from "@/utils/encoding";
import { formatCount } from "@/utils/count";
import DeviceBriefDialog from "./DeviceBriefDialog";
import { boundsFilter, compactCount, type MapArea } from "./mapModel";

const PAGE_SIZE = 100;
const SCROLL_ROOT = "[data-area-devices]";
const OPEN_DIALOG =
  'dialog[open], [role="dialog"][aria-modal="true"]:not([inert])';

interface AreaDevicesPanelProps {
  area: MapArea;
  folded: boolean;
  onToggleFold: () => void;
  onClose: () => void;
}

/**
 * The devices inside the area picked on the map, read by cursor as the list scrolls, each opening
 * a brief of the device over the map. An area that holds one device opens that device's brief
 * straight away, unless its first page holds other than one device or cannot be read, when the
 * list shows as for any area. That choice is made from the counts the panel mounts with and holds
 * for the life of the panel, whatever later counts or a later read of the area return. Folded,
 * the panel shows only its header and keeps the list, and where it was scrolled to, for unfolding;
 * the caller holds folded, so it outlives the panel. A brief opened straight away for a one-device
 * area opens whether or not the panel is folded. Escape closes the panel, folded or not, unless a
 * dialog or drawer is open, the brief or any other, where it closes that one alone, or a terminal or
 * recording window covers the map, where it leaves the panel alone.
 */
export default function AreaDevicesPanel({
  area,
  folded,
  onToggleFold,
  onClose,
}: AreaDevicesPanelProps) {
  const listId = useId();
  const footerId = useId();
  const [briefed, setBriefed] = useState<Device | null>(null);
  const windowShown = useWindowShown();
  const dialogOpenOnEscape = useRef(false);
  useEffect(() => {
    const recordOpenDialog = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        dialogOpenOnEscape.current = document.querySelector(OPEN_DIALOG) !== null;
      }
    };
    document.addEventListener("keydown", recordOpenDialog, true);
    return () => document.removeEventListener("keydown", recordOpenDialog, true);
  }, []);
  useEscapeKey(() => {
    if (!dialogOpenOnEscape.current) onClose();
  }, !windowShown);
  const filter = toBase64Json(boundsFilter(area.bounds));
  const query = { status: "accepted" as const, filter };

  const {
    data,
    isLoading,
    isPending,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: [...getDevicesQueryKey({ query }), "cursor"],
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam, signal }) => {
      const { data: devices, response } = await getDevices({
        query: {
          ...query,
          "page[size]": PAGE_SIZE,
          ...(pageParam ? { "page[after]": pageParam } : {}),
        },
        signal,
        throwOnError: true,
      });

      return {
        devices,
        next: response.headers.get("X-Next-Cursor") ?? undefined,
      };
    },
    getNextPageParam: (last) => last.next,
  });

  const sentinel = useLoadOnReach(
    () => {
      if (!isFetchingNextPage) void fetchNextPage();
    },
    Boolean(hasNextPage),
    SCROLL_ROOT,
  );

  const devices = data?.pages.flatMap((page) => page.devices) ?? [];
  const [pointView, setPointView] = useState<"choosing" | "list" | Device>(
    area.devices === 1 ? "choosing" : "list",
  );

  const choosingBrief = pointView === "choosing";

  if (choosingBrief && !isPending) {
    const firstPage = data?.pages[0].devices;
    setPointView(firstPage?.length === 1 ? firstPage[0] : "list");
  }

  if (choosingBrief) {
    return (
      <div className="pointer-events-none absolute inset-0 z-20 flex items-center justify-center">
        <Spinner size="xl" aria-label="Loading device" />
      </div>
    );
  }
  if (typeof pointView === "object") {
    return <DeviceBriefDialog device={pointView} onClose={onClose} />;
  }

  return (
    <aside
      className={cn(
        "absolute right-4 top-4 z-20 flex w-96 flex-col overflow-hidden rounded-xl border border-border bg-surface/95 shadow-float backdrop-blur animate-slide-up max-sm:left-4 max-sm:top-auto max-sm:bottom-4 max-sm:w-auto",
        !folded && "bottom-10 max-sm:h-1/2",
      )}
    >
      <header
        className={cn(
          "flex items-start gap-1 px-4 py-3",
          !folded && "border-b border-border",
        )}
      >
        <div className="mr-auto">
          <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
            This area
          </p>
          <p className="mt-1 text-sm font-semibold text-text-primary">
            {compactCount(area.devices)}{" "}
            {area.devices === 1 ? "device" : "devices"}
          </p>
          <p className="font-mono text-2xs text-text-secondary">
            <span className="text-accent-green">
              {compactCount(area.online)} online
            </span>
            {" · "}
            {compactCount(area.devices - area.online)} offline
          </p>
        </div>
        <IconButton
          aria-label="Devices in this area"
          aria-expanded={!folded}
          aria-controls={`${listId} ${footerId}`}
          variant="ghost"
          size="sm"
          onClick={onToggleFold}
        >
          <ChevronUpIcon
            className={cn(
              "h-4 w-4 transition-transform",
              folded && "rotate-180",
            )}
          />
        </IconButton>
        <IconButton
          aria-label="Close area"
          variant="ghost"
          size="sm"
          onClick={onClose}
        >
          <XMarkIcon className="h-4 w-4" />
        </IconButton>
      </header>

      <div
        id={listId}
        data-area-devices
        hidden={folded}
        className="min-h-0 flex-1 overflow-y-auto"
      >
        {isLoading && (
          <div className="flex h-full items-center justify-center">
            <Spinner />
          </div>
        )}
        {error && (
          <p className="px-4 py-6 text-sm text-accent-red">
            Could not load the devices in this area.
          </p>
        )}
        <ul>
          {devices.map((device) => (
            <li key={device.uid}>
              <button
                type="button"
                onClick={() => setBriefed(device)}
                className="flex w-full items-center gap-3 border-b border-border/60 px-4 py-2.5 text-left hover:bg-hover-medium"
              >
                <StatusDot online={device.online} size="sm" />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm text-text-primary">
                    {device.name}
                  </span>
                  <span className="block truncate font-mono text-2xs text-text-muted">
                    {device.info?.pretty_name || device.info?.platform}
                    {device.info?.version && ` · ${device.info.version}`}
                  </span>
                </span>
              </button>
            </li>
          ))}
        </ul>
        <div ref={sentinel} aria-hidden="true" />
        {isFetchingNextPage && (
          <div className="flex justify-center py-3">
            <Spinner size="sm" />
          </div>
        )}
      </div>

      <footer
        id={footerId}
        hidden={folded}
        className="border-t border-border px-4 py-2 text-2xs font-mono text-text-muted"
      >
        {formatCount(devices.length)} loaded
      </footer>

      {briefed && (
        <DeviceBriefDialog device={briefed} onClose={() => setBriefed(null)} />
      )}
    </aside>
  );
}
