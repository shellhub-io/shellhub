import { describe, expect, it } from "vitest";
import type { DeviceLocation } from "@/client";
import {
  boundsFilter,
  framingBounds,
  locationsBounds,
  mapBackend,
  worldSpanZoom,
} from "../mapModel";

function at(
  latitude: number,
  longitude: number,
  devices = 1,
  online = devices,
): DeviceLocation {
  return { latitude, longitude, devices, online };
}

describe("mapBackend", () => {
  it("draws the bundled outline for outline", () => {
    expect(mapBackend("outline", "dark")).toEqual({ kind: "outline" });
  });

  it.each([
    {
      theme: "dark" as const,
      styleUrl: "https://tiles.openfreemap.org/styles/dark",
    },
    {
      theme: "light" as const,
      styleUrl: "https://tiles.openfreemap.org/styles/positron",
    },
  ])("follows the $theme theme for openfreemap", ({ theme, styleUrl }) => {
    expect(mapBackend("openfreemap", theme)).toEqual({
      kind: "tiles",
      styleUrl,
    });
  });

  it("uses an operator's style URL for both themes", () => {
    const styleUrl = "https://maps.example.com/style.json";

    expect(mapBackend(styleUrl, "dark")).toEqual({ kind: "tiles", styleUrl });
    expect(mapBackend(styleUrl, "light")).toEqual({ kind: "tiles", styleUrl });
  });

  it("falls back to the outline for a value it does not know", () => {
    expect(mapBackend("", "dark")).toEqual({ kind: "outline" });
  });
});

describe("locationsBounds", () => {
  it("is undefined with no location", () => {
    expect(locationsBounds([])).toBeUndefined();
  });

  it("keeps every device of a small namespace in frame", () => {
    const locations = [
      at(46.5, 6.6),
      at(54.3, 10.1),
      at(51.4, 6.8),
      at(47.5, 8.5),
    ];

    expect(locationsBounds(locations)).toEqual([6.6, 46.5, 10.1, 54.3]);
  });

  it("leaves a few far-off devices out of a large namespace's frame", () => {
    const locations = [
      at(47.4, 8.5, 600),
      at(52.5, 13.4, 390),
      at(-23.5, -46.6, 5),
      at(35.7, 139.7, 5),
    ];

    expect(locationsBounds(locations)).toEqual([8.5, 47.4, 13.4, 52.5]);
  });
});

describe("boundsFilter", () => {
  it("selects the located devices strictly inside the bounds", () => {
    expect(boundsFilter([8.5, 47.3, 8.6, 47.4])).toEqual([
      { type: "operator", params: { name: "and" } },
      {
        type: "property",
        params: { name: "latitude", operator: "gt", value: 47.3 },
      },
      {
        type: "property",
        params: { name: "latitude", operator: "lt", value: 47.4 },
      },
      {
        type: "property",
        params: { name: "longitude", operator: "gt", value: 8.5 },
      },
      {
        type: "property",
        params: { name: "longitude", operator: "lt", value: 8.6 },
      },
      {
        type: "property",
        params: { name: "located", operator: "bool", value: true },
      },
    ]);
  });
});

describe("framingBounds", () => {
  it("keeps bounds that already span half a degree", () => {
    expect(framingBounds([6, 46, 15, 54])).toEqual([6, 46, 15, 54]);
  });

  it("widens a single location by a quarter degree on every side", () => {
    expect(framingBounds([8.54, 47.38, 8.54, 47.38])).toEqual([8.29, 47.13, 8.79, 47.63]);
  });
});

describe("worldSpanZoom", () => {
  it("is the zoom at which the world is exactly as wide as the map", () => {
    expect(worldSpanZoom(512)).toBe(0);
    expect(worldSpanZoom(1024)).toBe(1);
    expect(worldSpanZoom(2048)).toBe(2);
  });
});
