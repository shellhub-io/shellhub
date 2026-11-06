#!/usr/bin/env node

import { writeFile } from "node:fs/promises";

const NATURAL_EARTH_VERSION = "v5.1.2";
const SOURCE = `https://raw.githubusercontent.com/nvkelso/natural-earth-vector/${NATURAL_EARTH_VERSION}/geojson/ne_10m_populated_places_simple.geojson`;
const OUTPUT = new URL(
  "../apps/console/src/pages/device-map/cities-10m.json",
  import.meta.url,
);

const response = await fetch(SOURCE);
if (!response.ok) {
  console.error(`🚫 ERROR: ${SOURCE} answered ${response.status}`);
  process.exit(1);
}

const { features } = await response.json();

const round = (value) => Math.round(value * 1000) / 1000;

const cities = features
  .map(({ geometry, properties }) => ({
    longitude: round(geometry.coordinates[0]),
    latitude: round(geometry.coordinates[1]),
    name: properties.name,
    minZoom: properties.min_zoom,
    population: properties.pop_max,
  }))
  .sort((a, b) => a.minZoom - b.minZoom || b.population - a.population)
  .map(({ longitude, latitude, name, minZoom }) => [
    longitude,
    latitude,
    name,
    minZoom,
  ]);

await writeFile(OUTPUT, JSON.stringify(cities));

console.log(
  `Wrote ${cities.length} cities from Natural Earth ${NATURAL_EARTH_VERSION} to ${OUTPUT.pathname}`,
);
