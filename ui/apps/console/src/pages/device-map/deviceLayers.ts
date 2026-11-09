import type {
  ExpressionSpecification,
  FilterSpecification,
  GeoJSONSource,
} from "maplibre-gl";
import type { FeatureCollection, GeoJsonProperties, Point } from "geojson";
import type { DeviceLocation } from "@/client";
import {
  LABEL_THOUSAND,
  LABEL_UNLABELLED,
  ONLINE_COLOR_STOPS,
  RADIUS_FLOOR,
  RADIUS_RANGE,
  pointsBounds,
  type MapArea,
  type MapMode,
  type MapPalette,
} from "./mapModel";

/**
 * The locations as GeoJSON points, the shape a clustered source takes. Each carries its location
 * as properties: the counts, and the exact longitude and latitude, which a rendered feature's
 * geometry loses to tile precision.
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
      properties: { ...location },
    })),
  };
}

/**
 * The area a picked point or cluster of toFeatureCollection covers: its counts, and bounds around
 * the exact positions in its properties, a cluster's read from its leaves through source. It
 * rejects when source cannot read the leaves. A feature must carry the properties
 * toFeatureCollection writes, or a cluster's over them: one without them throws, and one without
 * a position gives NaN bounds.
 */
export async function areaOf(
  source: Pick<GeoJSONSource, "getClusterLeaves">,
  feature: { properties: GeoJsonProperties },
): Promise<MapArea> {
  const properties = feature.properties as DeviceLocation & {
    cluster?: boolean;
    cluster_id?: number;
    point_count?: number;
  };

  if (!properties.cluster || properties.cluster_id === undefined) {
    return {
      bounds: pointsBounds([[properties.longitude, properties.latitude]]),
      devices: properties.devices,
      online: properties.online,
    };
  }

  const leaves = await source.getClusterLeaves(
    properties.cluster_id,
    properties.point_count ?? Infinity,
    0,
  );

  return {
    bounds: pointsBounds(
      leaves.map((leaf) => {
        const { longitude, latitude } = leaf.properties as DeviceLocation;

        return [longitude, latitude];
      }),
    ),
    devices: properties.devices,
    online: properties.online,
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
