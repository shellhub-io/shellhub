import type { ExpressionSpecification, FilterSpecification } from "maplibre-gl";
import type { FeatureCollection, Point } from "geojson";
import type { DeviceLocation } from "@/client";
import {
  LABEL_THOUSAND,
  LABEL_UNLABELLED,
  ONLINE_COLOR_STOPS,
  RADIUS_FLOOR,
  RADIUS_RANGE,
  type MapMode,
  type MapPalette,
} from "./mapModel";

/**
 * The locations as GeoJSON points carrying their counts, the shape a clustered source takes.
 */
export function toFeatureCollection(
  locations: DeviceLocation[],
): FeatureCollection<Point> {
  return {
    type: "FeatureCollection",
    features: locations.map((location) => ({
      type: "Feature",
      geometry: {
        type: "Point",
        coordinates: [location.longitude, location.latitude],
      },
      properties: { devices: location.devices, online: location.online },
    })),
  };
}

function metric(mode: MapMode): ExpressionSpecification {
  switch (mode) {
    case "online":
      return ["get", "online"];
    case "offline":
      return ["-", ["get", "devices"], ["get", "online"]];
    default:
      return ["get", "devices"];
  }
}

/**
 * Keeps the points and clusters that count at least one device under mode.
 */
export function metricFilter(mode: MapMode): FilterSpecification {
  return [">", metric(mode), 0];
}

/**
 * A circle's radius in pixels over the feature's count under mode: RADIUS_FLOOR plus RADIUS_RANGE
 * times the square root of its share of total, so its area tracks its devices and the scale
 * follows the namespace. A small fleet draws circles as legible as a large one.
 */
export function radius(mode: MapMode, total: number): ExpressionSpecification {
  return [
    "+",
    RADIUS_FLOOR,
    ["*", RADIUS_RANGE, ["sqrt", ["/", metric(mode), Math.max(total, 1)]]],
  ];
}

/**
 * A circle's colour: under all, interpolated between ONLINE_COLOR_STOPS over the feature's share
 * online; under online, the online colour; under offline, the muted one.
 */
export function color(
  mode: MapMode,
  palette: MapPalette,
): ExpressionSpecification | string {
  if (mode === "online") return palette.online;
  if (mode === "offline") return palette.muted;

  return [
    "interpolate",
    ["linear"],
    ["/", ["get", "online"], ["max", ["get", "devices"], 1]],
    ...ONLINE_COLOR_STOPS.flatMap(([at, key]) => [at, palette[key]]),
  ] as ExpressionSpecification;
}

/**
 * A circle's label over the feature's count under mode: in thousands to one decimal from
 * LABEL_THOUSAND on, and nothing at LABEL_UNLABELLED or fewer.
 */
export function label(mode: MapMode): ExpressionSpecification {
  const count = metric(mode);

  return [
    "case",
    [">=", count, LABEL_THOUSAND],
    ["concat", ["to-string", ["/", ["round", ["/", count, 100]], 10]], "k"],
    [">", count, LABEL_UNLABELLED],
    ["to-string", count],
    "",
  ];
}
