import { useEffect, useRef, useState } from "react";
import {
  AttributionControl,
  Map as MapLibre,
  NavigationControl,
  setWorkerUrl,
  type GeoJSONSource,
  type MapLayerMouseEvent,
  type MapMouseEvent,
} from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import type { FeatureCollection, Polygon } from "geojson";
import workerUrl from "maplibre-gl/dist/maplibre-gl-worker.mjs?url";
import "./mapControls.css";
import MapTooltip, { type MapHover } from "./MapTooltip";
import { OUTLINE_FONT, outlineStyle } from "./outlineStyle";
import {
  CLUSTER_MAX_ZOOM,
  CLUSTER_RADIUS,
  FULL_PADDING,
  framingBounds,
  locationsBounds,
  MAP_MAX_ZOOM,
  modeTotal,
  paddingWithin,
  STILL_PADDING,
  readPalette,
  worldSpanZoom,
  type Bounds,
  type LocationsMapProps,
  type MapBackend,
  type MapMode,
  type MapPalette,
  type Padding,
} from "./mapModel";
import {
  areaOf,
  color,
  label,
  metricFilter,
  radius,
  toFeatureCollection,
} from "./deviceLayers";

setWorkerUrl(workerUrl);

const SOURCE = "locations";
const SELECTION_SOURCE = "selection";
const CIRCLES = "location-circles";
const LABELS = "location-labels";
const SELECTION = "selection-outline";

interface MapLibreMapProps extends LocationsMapProps {
  backend: MapBackend;
}

const TILES_FONT = ["Noto Sans Bold"];

function selectionData(bounds?: Bounds): FeatureCollection<Polygon> {
  if (!bounds) return { type: "FeatureCollection", features: [] };

  const [west, south, east, north] = bounds;

  return {
    type: "FeatureCollection",
    features: [
      {
        type: "Feature",
        properties: {},
        geometry: {
          type: "Polygon",
          coordinates: [
            [
              [west, south],
              [east, south],
              [east, north],
              [west, north],
              [west, south],
            ],
          ],
        },
      },
    ],
  };
}

function applyMode(
  map: MapLibre,
  mode: MapMode,
  palette: MapPalette,
  total: number,
) {
  map.setFilter(CIRCLES, metricFilter(mode));
  map.setFilter(LABELS, metricFilter(mode));
  map.setPaintProperty(CIRCLES, "circle-radius", radius(mode, total));
  map.setPaintProperty(CIRCLES, "circle-color", color(mode, palette));
  map.setLayoutProperty(LABELS, "text-field", label(mode));
}

function addLayers(map: MapLibre, palette: MapPalette, font: string[]) {
  map.addSource(SOURCE, {
    type: "geojson",
    data: toFeatureCollection([]),
    cluster: true,
    clusterRadius: CLUSTER_RADIUS,
    clusterMaxZoom: CLUSTER_MAX_ZOOM,
    clusterProperties: {
      devices: ["+", ["get", "devices"]],
      online: ["+", ["get", "online"]],
    },
  });

  map.addSource(SELECTION_SOURCE, { type: "geojson", data: selectionData() });

  map.addLayer({
    id: SELECTION,
    type: "line",
    source: SELECTION_SOURCE,
    paint: {
      "line-color": palette.selection,
      "line-width": 1.5,
      "line-dasharray": [3, 2],
    },
  });

  map.addLayer({
    id: CIRCLES,
    type: "circle",
    source: SOURCE,
    paint: {
      "circle-opacity": 0.82,
      "circle-stroke-color": palette.halo,
      "circle-stroke-width": 1.5,
      "circle-stroke-opacity": 0.9,
    },
  });

  map.addLayer({
    id: LABELS,
    type: "symbol",
    source: SOURCE,
    layout: {
      "text-font": font,
      "text-size": 11,
      "text-allow-overlap": true,
    },
    paint: {
      "text-color": palette.text,
      "text-halo-color": palette.halo,
      "text-halo-width": 1.2,
    },
  });
}

function frame(
  map: MapLibre,
  [west, south, east, north]: Bounds,
  padding: Padding,
  duration: number,
) {
  const { clientWidth, clientHeight } = map.getContainer();
  map.fitBounds(
    [
      [west, south],
      [east, north],
    ],
    {
      padding: paddingWithin(padding, clientWidth, clientHeight),
      maxZoom: MAP_MAX_ZOOM,
      duration,
    },
  );
}

function collapseAttribution(map: MapLibre) {
  const attribution = map
    .getContainer()
    .querySelector("details.maplibregl-ctrl-attrib");
  attribution?.removeAttribute("open");
  attribution?.classList.remove("maplibregl-compact-show");
}

/**
 * The device map drawn by MapLibre, over the vector tiles of a style URL or over the bundled
 * outline, clustering with MapLibre's own clusters either way. It behaves as LocationsMapProps
 * describes.
 */
export default function MapLibreMap({
  locations,
  mode,
  selected,
  onSelect,
  navigable = false,
  areaPadding = FULL_PADDING,
  backend,
}: MapLibreMapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<MapLibre | null>(null);
  const paletteRef = useRef<MapPalette | null>(null);
  const framedRef = useRef(false);
  const onSelectRef = useRef(onSelect);
  const areaPaddingRef = useRef(areaPadding);
  const [ready, setReady] = useState(false);
  const [hover, setHover] = useState<MapHover | null>(null);
  const clickable = onSelect !== undefined;
  const total = modeTotal(locations, mode);
  const styleUrl = backend.kind === "tiles" ? backend.styleUrl : undefined;

  useEffect(() => {
    onSelectRef.current = onSelect;
    areaPaddingRef.current = areaPadding;
  });

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const palette = readPalette();
    paletteRef.current = palette;
    framedRef.current = false;

    const outline = styleUrl === undefined;
    const map = new MapLibre({
      container,
      style: styleUrl ?? outlineStyle(palette),
      renderWorldCopies: false,
      center: [0, 20],
      zoom: worldSpanZoom(container.clientWidth),
      maxZoom: MAP_MAX_ZOOM,
      attributionControl: false,
      interactive: clickable || navigable,
      dragPan: navigable,
      scrollZoom: navigable,
      boxZoom: navigable,
      doubleClickZoom: navigable,
      keyboard: navigable,
      touchZoomRotate: navigable,
      touchPitch: navigable,
      dragRotate: false,
      pitchWithRotate: false,
    });
    map.addControl(new AttributionControl({ compact: true }), "bottom-left");
    mapRef.current = map;

    map.on("load", () => {
      collapseAttribution(map);
      addLayers(map, palette, outline ? OUTLINE_FONT : TILES_FONT);
      setReady(true);
    });

    const restingCursor = navigable ? "" : "default";
    map.getCanvas().style.cursor = restingCursor;

    if (navigable) {
      const keepTheWorldSpanningTheMap = () =>
        map.setMinZoom(worldSpanZoom(container.clientWidth));
      keepTheWorldSpanningTheMap();
      map.on("resize", keepTheWorldSpanningTheMap);
      map.touchZoomRotate.disableRotation();
      map.addControl(
        new NavigationControl({ showCompass: false }),
        "bottom-left",
      );
    }

    if (clickable) {
      map.on("mousemove", CIRCLES, (event: MapLayerMouseEvent) => {
        const feature = event.features?.[0];
        if (!feature) return;

        map.getCanvas().style.cursor = "pointer";
        const properties = feature.properties as {
          devices: number;
          online: number;
          cluster?: boolean;
        };
        setHover({
          x: event.point.x,
          y: event.point.y,
          devices: properties.devices,
          online: properties.online,
          cluster: Boolean(properties.cluster),
        });
      });

      map.on("mouseleave", CIRCLES, () => {
        map.getCanvas().style.cursor = restingCursor;
        setHover(null);
      });

      map.on("click", (event: MapMouseEvent) => {
        if (!map.getLayer(CIRCLES)) return;

        const [feature] = map.queryRenderedFeatures(event.point, {
          layers: [CIRCLES],
        });
        if (!feature) {
          onSelectRef.current?.(null);
          return;
        }

        const source = map.getSource<GeoJSONSource>(SOURCE);
        if (!source) return;

        void areaOf(source, feature).then((area) => {
          if (navigable) {
            frame(map, area.bounds, areaPaddingRef.current, 700);
          }
          setHover(null);
          onSelectRef.current?.(area);
        });
      });
    }

    return () => {
      setReady(false);
      setHover(null);
      mapRef.current = null;
      map.remove();
    };
  }, [styleUrl, clickable, navigable]);

  useEffect(() => {
    const map = mapRef.current;
    if (!ready || !map) return;

    void map
      .getSource<GeoJSONSource>(SOURCE)
      ?.setData(toFeatureCollection(locations));
  }, [ready, locations]);

  useEffect(() => {
    const map = mapRef.current;
    if (!ready || !map || framedRef.current) return;

    const located = locationsBounds(locations);
    const bounds = selected ?? (located && framingBounds(located));
    if (!bounds) return;

    framedRef.current = true;
    const padding = navigable ? FULL_PADDING : STILL_PADDING;
    frame(map, bounds, selected ? areaPaddingRef.current : padding, 0);
  }, [ready, locations, selected, navigable]);

  useEffect(() => {
    const map = mapRef.current;
    const palette = paletteRef.current;
    if (!ready || !map || !palette) return;

    applyMode(map, mode, palette, total);
  }, [ready, mode, total]);

  useEffect(() => {
    const map = mapRef.current;
    if (!ready || !map) return;

    void map
      .getSource<GeoJSONSource>(SELECTION_SOURCE)
      ?.setData(selectionData(selected));
  }, [ready, selected]);

  return (
    <div className="absolute inset-0">
      <div ref={containerRef} className="h-full w-full" />
      {hover && <MapTooltip hover={hover} />}
    </div>
  );
}
