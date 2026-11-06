import { describe, expect, it } from "vitest";
import {
  createPropertyExpression,
  featureFilter,
  latest,
  type StylePropertySpecification,
} from "@maplibre/maplibre-gl-style-spec";
import type {
  GeoJSONSourceSpecification,
  SymbolLayerSpecification,
} from "maplibre-gl";
import type {
  Feature,
  FeatureCollection,
  MultiPolygon,
  Point,
  Polygon,
} from "geojson";
import { mockMapPalette } from "@/tests/factories";
import { outlineStyle } from "../outlineStyle";

const palette = mockMapPalette();

const style = outlineStyle(palette);

function source(id: string): FeatureCollection<Point> {
  const spec = style.sources[id] as GeoJSONSourceSpecification;

  return spec.data as FeatureCollection<Point>;
}

function symbolLayer(id: string): SymbolLayerSpecification {
  const layer = style.layers.find((candidate) => candidate.id === id);
  if (layer?.type !== "symbol") throw new Error(`no symbol layer ${id}`);

  return layer;
}

function named(id: string, name: string): Feature<Point> {
  const feature = source(id).features.find(
    (candidate) => candidate.properties?.name === name,
  );
  if (!feature) throw new Error(`no ${name} in ${id}`);

  return feature;
}

function shownAt(layerId: string, feature: Feature<Point>, zoom: number) {
  const filter = featureFilter(
    symbolLayer(layerId).filter,
    "filter",
  );

  return filter.filter({ zoom }, { type: 1, properties: feature.properties ?? {} });
}

describe("outlineStyle", () => {
  it("asks for nothing over the network", () => {
    expect(style.glyphs).toBeUndefined();
    expect(style.sprite).toBeUndefined();
    expect(JSON.stringify(style)).not.toMatch(/https?:\/\//);
  });

  it("draws every country", () => {
    expect(source("countries").features.length).toBeGreaterThan(150);
  });

  it("draws countries across the antimeridian without a stripe over the map", () => {
    const countries = (style.sources.countries as GeoJSONSourceSpecification)
      .data as FeatureCollection<Polygon | MultiPolygon>;

    const edgesOver180 = (name: string) => {
      const geometry = countries.features.find(
        (candidate) => candidate.properties?.name === name,
      )?.geometry;
      if (!geometry) throw new Error(`no ${name}`);
      const polygons =
        geometry.type === "Polygon" ? [geometry.coordinates] : geometry.coordinates;

      return polygons
        .flat()
        .flatMap((ring) => ring.slice(1).map((point, i) => [ring[i], point]))
        .filter(([from, to]) => Math.abs(to[0] - from[0]) > 180);
    };

    expect(edgesOver180("Russia")).toEqual([]);
    expect(edgesOver180("Fiji")).toEqual([]);

    const [antarcticaSouthernEdge, ...others] = edgesOver180("Antarctica");
    expect(others).toEqual([]);
    expect(antarcticaSouthernEdge[0][1]).toBe(antarcticaSouthernEdge[1][1]);
  });

  it("names a country inside its widest landmass", () => {
    const [longitude, latitude] = named("country-labels", "France").geometry.coordinates;

    expect(longitude).toBeGreaterThan(-5);
    expect(longitude).toBeLessThan(8);
    expect(latitude).toBeGreaterThan(42);
    expect(latitude).toBeLessThan(51);
  });

  it("names a large country from afar and a small one only closer in", () => {
    expect(shownAt("country-labels", named("country-labels", "Brazil"), 1)).toBe(true);
    expect(shownAt("country-labels", named("country-labels", "Luxembourg"), 1)).toBe(false);
    expect(shownAt("country-labels", named("country-labels", "Luxembourg"), 8)).toBe(true);
  });

  it("names a metropolis from afar and a small town only closer in", () => {
    expect(shownAt("city-labels", named("cities", "Berlin"), 2)).toBe(true);
    expect(shownAt("city-labels", named("cities", "Aarau"), 5)).toBe(false);
    expect(shownAt("city-labels", named("cities", "Aarau"), 8)).toBe(true);
  });

  it("places more important cities first when names collide", () => {
    const parsed = createPropertyExpression(
      symbolLayer("city-labels").layout?.["symbol-sort-key"],
      "layers[0].layout.symbol-sort-key",
      latest.layout_symbol["symbol-sort-key"] as unknown as StylePropertySpecification,
    );
    if (parsed.result === "error") throw new Error("city-labels has no sort key");
    const sortKey = (city: string) =>
      parsed.value.evaluate(
        { zoom: 8 },
        { type: 1, properties: named("cities", city).properties ?? {} },
      ) as number;

    expect(sortKey("Berlin")).toBeLessThan(sortKey("Aarau"));
  });
});
