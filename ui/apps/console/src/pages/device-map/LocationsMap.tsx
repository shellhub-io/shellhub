import { lazy, Suspense, useState } from "react";
import { getConfig } from "@/env";
import { useThemeStore } from "@/stores/themeStore";
import GeoIPCredit from "./GeoIPCredit";
import { mapBackend, type LocationsMapProps } from "./mapModel";

const MapLibreMap = lazy(() => import("./MapLibreMap"));

function browserDrawsMaps(): boolean {
  return document.createElement("canvas").getContext("webgl2") !== null;
}

/**
 * The device map, drawn by MapLibre over what SHELLHUB_MAP_STYLE picks: the bundled country
 * outline and cities, or the vector tiles of a style. MapLibre loads with the map, and the map
 * remounts on a theme change so it redraws in the new theme's colours. It behaves as
 * LocationsMapProps describes whatever it draws over. A browser without WebGL 2 gets a notice in
 * the map's place, and MapLibre is never loaded.
 */
export default function LocationsMap(props: LocationsMapProps) {
  const theme = useThemeStore((s) => s.theme);
  const [drawsMaps] = useState(browserDrawsMaps);
  const backend = mapBackend(getConfig().mapStyle, theme);
  const key =
    backend.kind === "tiles"
      ? `${backend.styleUrl}:${theme}`
      : `outline:${theme}`;

  if (!drawsMaps) {
    return (
      <p className="absolute inset-0 flex items-center justify-center px-6 text-center text-sm text-text-muted">
        This browser cannot draw the map: it needs WebGL 2.
      </p>
    );
  }

  return (
    <>
      <Suspense fallback={null}>
        <MapLibreMap key={key} backend={backend} {...props} />
      </Suspense>
      <GeoIPCredit source={props.source} />
    </>
  );
}
