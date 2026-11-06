import type {
  ExpressionSpecification,
  StyleSpecification,
} from "maplibre-gl";
import type {
  Feature,
  FeatureCollection,
  MultiPolygon,
  Point,
  Polygon,
  Position,
} from "geojson";
import polylabel from "polylabel";
import { feature } from "topojson-client";
import type { GeometryCollection, Topology } from "topojson-specification";
import countries110m from "world-atlas/countries-110m.json";
import type { MapPalette } from "./mapModel";
import cities10m from "./cities-10m.json";

/**
 * The font the outline draws its names in, from the fonts the page already has, since the
 * outline asks for no glyphs.
 */
export const OUTLINE_FONT = ["IBM Plex Sans", "sans-serif"];

type City = [longitude: number, latitude: number, name: string, minZoom: number];

const NATURAL_EARTH_256_PX_TILE_ZOOM_OFFSET = 1;
const COUNTRY_LABEL_BASE_ZOOM = 2;
const COUNTRY_LABEL_MAX_ZOOM = 8;
const LABEL_POINT_PRECISION = 0.1;

const topology = countries110m as unknown as Topology<{
  countries: GeometryCollection<{ name: string }>;
}>;

function antimeridianJumps(ring: Position[]): number {
  return ring.filter((point, i) => i > 0 && Math.abs(point[0] - ring[i - 1][0]) > 180)
    .length;
}

function unwrapRing(ring: Position[]): Position[] {
  const jumps = antimeridianJumps(ring);
  const goesAroundTheGlobe = jumps % 2 === 1;
  if (jumps === 0 || goesAroundTheGlobe) return ring;

  return ring.map(([longitude, latitude]) => [
    longitude < 0 ? longitude + 360 : longitude,
    latitude,
  ]);
}

function unwrapGeometry(geometry: Polygon | MultiPolygon): Polygon | MultiPolygon {
  if (geometry.type === "Polygon") {
    return { ...geometry, coordinates: geometry.coordinates.map(unwrapRing) };
  }

  return {
    ...geometry,
    coordinates: geometry.coordinates.map((rings) => rings.map(unwrapRing)),
  };
}

const COUNTRIES_AS_STORED = feature(
  topology,
  topology.objects.countries,
) as FeatureCollection<Polygon | MultiPolygon, { name: string }>;

const COUNTRIES: FeatureCollection<Polygon | MultiPolygon, { name: string }> = {
  ...COUNTRIES_AS_STORED,
  features: COUNTRIES_AS_STORED.features.map((country) => ({
    ...country,
    geometry: unwrapGeometry(country.geometry),
  })),
};

function labelPoint(geometry: Polygon | MultiPolygon) {
  const polygons =
    geometry.type === "Polygon" ? [geometry.coordinates] : geometry.coordinates;

  return polygons
    .map((rings) => polylabel(rings as [number, number][][], LABEL_POINT_PRECISION))
    .reduce((widest, point) => (point.distance > widest.distance ? point : widest));
}

const COUNTRY_LABELS: FeatureCollection<Point> = {
  type: "FeatureCollection",
  features: COUNTRIES.features.flatMap((country): Feature<Point>[] => {
    if (!country.geometry || !country.properties?.name) return [];

    const point = labelPoint(country.geometry);
    const minZoom = Math.min(
      Math.max(COUNTRY_LABEL_BASE_ZOOM - Math.log2(point.distance), 0),
      COUNTRY_LABEL_MAX_ZOOM,
    );

    return [
      {
        type: "Feature",
        geometry: { type: "Point", coordinates: [point[0], point[1]] },
        properties: { name: country.properties.name, min_zoom: minZoom },
      },
    ];
  }),
};

const CITIES: FeatureCollection<Point> = {
  type: "FeatureCollection",
  features: (cities10m as City[]).map(
    ([longitude, latitude, name, minZoom], rank): Feature<Point> => ({
      type: "Feature",
      geometry: { type: "Point", coordinates: [longitude, latitude] },
      properties: { name, min_zoom: minZoom - NATURAL_EARTH_256_PX_TILE_ZOOM_OFFSET, rank },
    }),
  ),
};

const reachedMinZoom: ExpressionSpecification = [
  ">=",
  ["zoom"],
  ["get", "min_zoom"],
];

/**
 * The style the outline map draws: the world's countries at 1:110m from world-atlas, each named
 * at the most inner point of its widest landmass once it is large enough on screen, and Natural
 * Earth's 1:10m cities, each named from the zoom Natural Earth sets for it, the more important
 * first where names collide. Everything is bundled: the style names no URL and no glyphs, so it
 * asks nothing of the network and draws its names with the fonts the page has.
 */
export function outlineStyle(palette: MapPalette): StyleSpecification {
  return {
    version: 8,
    sources: {
      countries: { type: "geojson", data: COUNTRIES },
      "country-labels": { type: "geojson", data: COUNTRY_LABELS },
      cities: { type: "geojson", data: CITIES },
    },
    layers: [
      {
        id: "background",
        type: "background",
        paint: { "background-color": palette.background },
      },
      {
        id: "country-fill",
        type: "fill",
        source: "countries",
        paint: { "fill-color": palette.land },
      },
      {
        id: "country-line",
        type: "line",
        source: "countries",
        paint: { "line-color": palette.border, "line-width": 0.75 },
      },
      {
        id: "country-labels",
        type: "symbol",
        source: "country-labels",
        filter: reachedMinZoom,
        layout: {
          "text-field": ["get", "name"],
          "text-font": OUTLINE_FONT,
          "text-size": 10,
          "text-transform": "uppercase",
          "text-letter-spacing": 0.15,
          "symbol-sort-key": ["get", "min_zoom"],
        },
        paint: {
          "text-color": palette.label,
          "text-halo-color": palette.background,
          "text-halo-width": 1,
        },
      },
      {
        id: "city-dots",
        type: "circle",
        source: "cities",
        filter: reachedMinZoom,
        paint: { "circle-radius": 2, "circle-color": palette.label },
      },
      {
        id: "city-labels",
        type: "symbol",
        source: "cities",
        filter: reachedMinZoom,
        layout: {
          "text-field": ["get", "name"],
          "text-font": OUTLINE_FONT,
          "text-size": 11,
          "text-anchor": "left",
          "text-offset": [0.6, 0],
          "symbol-sort-key": ["get", "rank"],
        },
        paint: {
          "text-color": palette.label,
          "text-halo-color": palette.background,
          "text-halo-width": 1,
        },
      },
    ],
  };
}
