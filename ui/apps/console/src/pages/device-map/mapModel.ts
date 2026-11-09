import type { DeviceLocation } from "@/client";
import type { GeoIPSource } from "@/types/geoip";
import { propertyFilter } from "@/utils/encoding";
import { formatCount } from "@/utils/count";
import type { AppTheme } from "@/stores/themeStore";

/**
 * Which devices the map counts: every located device, only the online ones, or only the offline
 * ones. Switching it redraws the clusters already loaded, without another request.
 */
export type MapMode = "all" | "online" | "offline";

/**
 * A rectangle of the map, west, south, east and north, in degrees.
 */
export type Bounds = [number, number, number, number];

/**
 * The part of the map a click picked: its bounds and the devices counted inside it.
 */
export interface MapArea {
  bounds: Bounds;
  devices: number;
  online: number;
}

const SERVER_ROUNDING_DECIMALS = 2;
const EDGE_MARGIN = 0.0001;
const LOCATION_HALF_STEP = 0.5 * 10 ** -SERVER_ROUNDING_DECIMALS + EDGE_MARGIN;

/**
 * What every map backend takes: the locations, the GeoIP database that placed them, which devices to count, the area picked, and
 * onSelect for a click on a point or cluster. Without onSelect the map has no hover or clicks.
 * Without navigable it has no panning, zooming or controls. Either way it opens once on the picked
 * area when there is one, or else on where most devices are. It frames the picked area it opens
 * on, and a navigable map each area picked, with areaPadding as it is when the map first frames or
 * at the pick, or the full map's padding when none is given. Any framing whose padding leaves less
 * than 64 pixels either way to frame in uses 20 pixels on every side instead, as paddingWithin does.
 */
export interface LocationsMapProps {
  locations: DeviceLocation[];
  source?: GeoIPSource;
  mode: MapMode;
  selected?: Bounds;
  onSelect?: (area: MapArea) => void;
  navigable?: boolean;
  areaPadding?: Padding;
}

/**
 * How far, in screen pixels, points merge into one cluster, on both backends.
 */
export const CLUSTER_RADIUS = 48;

/**
 * The zoom level past which points no longer cluster, on both backends.
 */
export const CLUSTER_MAX_ZOOM = 11;

/**
 * The closest the map zooms in, on both backends.
 */
export const MAP_MAX_ZOOM = 13;

const WORLD_WIDTH_AT_ZOOM_ZERO = 512;

/**
 * The zoom at which the whole world spans a map width pixels wide: the furthest the full map
 * zooms out.
 */
export function worldSpanZoom(width: number): number {
  return Math.log2(Math.max(width, 1) / WORLD_WIDTH_AT_ZOOM_ZERO);
}

/**
 * The space, in pixels, kept clear on each side of a viewport when framing.
 */
export interface Padding {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

/**
 * The padding when framing the devices on the still dashboard map, clear of the figures strip.
 */
export const STILL_PADDING: Padding = {
  top: 110,
  right: 40,
  bottom: 40,
  left: 40,
};

/**
 * The padding when framing the devices on the full map, clear of the side panel.
 */
export const FULL_PADDING: Padding = {
  top: 60,
  right: 60,
  bottom: 60,
  left: 340,
};

/**
 * Which of the full map's panels are folded. A folded summary card keeps its Dashboard link, its
 * title row with the online count, its mode switch and any load error or no-position notice; a
 * folded area panel keeps its header.
 */
export interface FoldedPanels {
  summary: boolean;
  area: boolean;
}

/**
 * The padding when framing a picked area on the full map, clear of the summary card on the left
 * and the area panel on the right. An open panel keeps its whole side clear. A folded one keeps
 * clear only the top, deep enough for what stays of it: the area panel's header, or the summary's
 * Dashboard link, title row and mode switch, the deeper of the two when both are folded. The
 * summary's load error or no-position notice is not cleared: either shows only when the map holds
 * nothing to pick.
 */
export function areaPadding(folded: FoldedPanels): Padding {
  return {
    top: Math.max(80, folded.area ? 110 : 0, folded.summary ? 160 : 0),
    right: folded.area ? 80 : 420,
    bottom: 80,
    left: folded.summary ? 80 : 340,
  };
}

const NARROW_PADDING: Padding = { top: 20, right: 20, bottom: 20, left: 20 };
const MIN_FRAME = 64;

/**
 * The padding to frame with in a map width by height pixels: the padding given when it leaves at
 * least 64 pixels each way to frame in, or else 20 pixels on every side. A map too narrow for its
 * panels still frames what it is asked to, under the panels.
 */
export function paddingWithin(
  padding: Padding,
  width: number,
  height: number,
): Padding {
  const roomy =
    width - padding.left - padding.right >= MIN_FRAME &&
    height - padding.top - padding.bottom >= MIN_FRAME;

  return roomy ? padding : NARROW_PADDING;
}

const OPENFREEMAP_STYLES: Record<AppTheme, string> = {
  dark: "https://tiles.openfreemap.org/styles/dark",
  light: "https://tiles.openfreemap.org/styles/positron",
};

/**
 * What draws the map: the bundled country outline, or vector tiles from a MapLibre style.
 */
export type MapBackend =
  { kind: "outline" } | { kind: "tiles"; styleUrl: string };

/**
 * The backend SHELLHUB_MAP_STYLE picks under theme: openfreemap follows the theme, an http(s) URL
 * is the operator's style for both themes, and anything else, outline included, is the outline.
 */
export function mapBackend(style: string, theme: AppTheme): MapBackend {
  if (style === "openfreemap") {
    return { kind: "tiles", styleUrl: OPENFREEMAP_STYLES[theme] };
  }

  if (/^https?:\/\//.test(style)) return { kind: "tiles", styleUrl: style };

  return { kind: "outline" };
}

/**
 * The radius, in pixels, of a circle holding a single device of a namespace too large to tell it
 * from none.
 */
export const RADIUS_FLOOR = 8;

/**
 * How many pixels a circle holding every device grows past RADIUS_FLOOR.
 */
export const RADIUS_RANGE = 40;

/**
 * Where a circle's colour sits for the share of its devices online: offline at half or less,
 * partial at four in five, online at all of them.
 */
export const ONLINE_COLOR_STOPS = [
  [0.5, "offline"],
  [0.8, "partial"],
  [1, "online"],
] as const satisfies readonly (readonly [number, keyof MapPalette])[];

/**
 * The count from which a circle's label reads in thousands, on both backends.
 */
export const LABEL_THOUSAND = 1000;

/**
 * The largest count a circle carries no label for, on both backends.
 */
export const LABEL_UNLABELLED = 1;

const FRAMING_MIN_SPAN = 0.5;
const FRAMING_WIDENING = 0.25;

/**
 * The bounds a map frames for bounds: bounds themselves when they span half a degree or more,
 * and otherwise bounds widened by a quarter degree on every side, so a single location does not
 * zoom in to the street. Both backends frame through it, so a namespace opens at the same zoom
 * whichever draws it.
 */
export function framingBounds([west, south, east, north]: Bounds): Bounds {
  if (Math.max(east - west, north - south) >= FRAMING_MIN_SPAN) {
    return [west, south, east, north];
  }

  return [
    west - FRAMING_WIDENING,
    south - FRAMING_WIDENING,
    east + FRAMING_WIDENING,
    north + FRAMING_WIDENING,
  ];
}

function countUnder(
  mode: MapMode,
  { devices, online }: { devices: number; online: number },
): number {
  if (mode === "online") return online;
  if (mode === "offline") return devices - online;

  return devices;
}

/**
 * How many devices the whole map counts under mode, the most any one circle can hold.
 */
export function modeTotal(locations: DeviceLocation[], mode: MapMode): number {
  return locations.reduce(
    (sum, location) => sum + countUnder(mode, location),
    0,
  );
}

function weightedQuantile(
  locations: DeviceLocation[],
  coordinate: "latitude" | "longitude",
  total: number,
  share: number,
): number {
  const sorted = [...locations].sort((a, b) => a[coordinate] - b[coordinate]);
  let covered = 0;
  for (const location of sorted) {
    covered += location.devices;
    if (covered >= total * share) return location[coordinate];
  }

  return sorted[sorted.length - 1][coordinate];
}

/**
 * The bounds of where most devices are: on each axis, the band between the devices at trim and
 * 1 - trim of the way along it, so a handful of far-off devices does not zoom the map out to the
 * whole world. Undefined when there is no location.
 */
export function locationsBounds(
  locations: DeviceLocation[],
  trim = 0.025,
): Bounds | undefined {
  if (locations.length === 0) return undefined;

  const total = modeTotal(locations, "all");

  return [
    weightedQuantile(locations, "longitude", total, trim),
    weightedQuantile(locations, "latitude", total, trim),
    weightedQuantile(locations, "longitude", total, 1 - trim),
    weightedQuantile(locations, "latitude", total, 1 - trim),
  ];
}

/**
 * The bounds of a set of rounded points, widened by half a rounding step on every side.
 */
export function pointsBounds(points: [number, number][]): Bounds {
  const [west, south, east, north] = points.reduce<Bounds>(
    ([w, s, e, n], [lng, lat]) => [
      Math.min(w, lng),
      Math.min(s, lat),
      Math.max(e, lng),
      Math.max(n, lat),
    ],
    [Infinity, Infinity, -Infinity, -Infinity],
  );

  return [
    west - LOCATION_HALF_STEP,
    south - LOCATION_HALF_STEP,
    east + LOCATION_HALF_STEP,
    north + LOCATION_HALF_STEP,
  ];
}

const BOUNDS_DECIMALS = 4;

/**
 * Bounds as the area URL parameter carries them: west, south, east and north joined by commas,
 * each rounded to four decimals. Rounding moves an edge by less than the margin pointsBounds
 * leaves around a location, so the bounds still hold the locations they were drawn around.
 */
export function formatBounds(bounds: Bounds): string {
  return bounds.map((edge) => Number(edge.toFixed(BOUNDS_DECIMALS))).join(",");
}

/**
 * The bounds an area URL parameter carries, or undefined when value is missing, holds other than
 * four numbers, has its west not before its east or its south not below its north, or has a
 * latitude outside -90 to 90, which MapLibre throws on when it frames the bounds.
 */
export function parseBounds(value: string | null): Bounds | undefined {
  const edges = value
    ?.split(",")
    .map((edge) => (edge.trim() === "" ? NaN : Number(edge)));
  if (edges?.length !== 4 || !edges.every(Number.isFinite)) return undefined;

  const [west, south, east, north] = edges;

  const ordered = west < east && south < north;
  const onEarth = south >= -90 && north <= 90;

  return ordered && onEarth ? [west, south, east, north] : undefined;
}

/**
 * The area of bounds over locations: the devices of every location strictly inside bounds, the
 * same locations boundsFilter lists.
 */
export function areaWithin(
  locations: DeviceLocation[],
  bounds: Bounds,
): MapArea {
  const [west, south, east, north] = bounds;
  const inside = locations.filter(
    ({ latitude, longitude }) =>
      latitude > south &&
      latitude < north &&
      longitude > west &&
      longitude < east,
  );

  return {
    bounds,
    devices: modeTotal(inside, "all"),
    online: modeTotal(inside, "online"),
  };
}

/**
 * The device list filter selecting the located devices inside bounds, as the API reads it. A device
 * with no position is kept at 0,0, so bounds around that point would otherwise list every one.
 */
export function boundsFilter([west, south, east, north]: Bounds) {
  return [
    { type: "operator", params: { name: "and" } },
    propertyFilter("latitude", "gt", south),
    propertyFilter("latitude", "lt", north),
    propertyFilter("longitude", "gt", west),
    propertyFilter("longitude", "lt", east),
    propertyFilter("located", "bool", true),
  ];
}

/**
 * The palette the map draws with, resolved from the design-system tokens.
 */
export interface MapPalette {
  online: string;
  partial: string;
  offline: string;
  muted: string;
  halo: string;
  text: string;
  selection: string;
  background: string;
  land: string;
  border: string;
  label: string;
}

function tokenColor(className: string): string {
  const probe = document.createElement("span");
  probe.className = className;
  document.body.appendChild(probe);
  const value = getComputedStyle(probe).color;
  probe.remove();

  return value;
}

/**
 * The colours the map draws with, read from the design-system tokens as the theme stands at the
 * call. It briefly attaches a probe element to the document to resolve them, so it runs only in
 * the browser, and a theme change takes a new call: the maps remount on one to pick it up.
 */
export function readPalette(): MapPalette {
  return {
    online: tokenColor("text-accent-green"),
    partial: tokenColor("text-accent-yellow"),
    offline: tokenColor("text-accent-red"),
    muted: tokenColor("text-text-muted"),
    halo: tokenColor("text-background"),
    text: tokenColor("text-text-primary"),
    selection: tokenColor("text-primary"),
    background: tokenColor("text-background"),
    land: tokenColor("text-card"),
    border: tokenColor("text-border-light"),
    label: tokenColor("text-text-secondary"),
  };
}

/**
 * A count in thousands from ten thousand on, grouped otherwise.
 */
export function compactCount(count: number): string {
  if (count >= 10_000) return `${(count / 1000).toFixed(1)}k`;

  return formatCount(count);
}
