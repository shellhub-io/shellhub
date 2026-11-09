import { afterEach, describe, expect, it, vi } from "vitest";
import { render, waitFor } from "@testing-library/react";
import MapLibreMap from "../MapLibreMap";

type Listener = (event: { point: { x: number; y: number } }) => void;

const maps = vi.hoisted(() => {
  class FakeMap {
    clickListeners: Listener[] = [];
    rendered: unknown[] = [];
    touchZoomRotate = { disableRotation: () => {} };
    queryRenderedFeatures = () => this.rendered;

    constructor() {
      created.push(this);
    }

    on(type: string, layerOrListener: string | Listener) {
      if (type === "click" && typeof layerOrListener === "function") {
        this.clickListeners.push(layerOrListener);
      }
    }

    clickAt(rendered: unknown[]) {
      this.rendered = rendered;
      for (const listener of this.clickListeners) {
        listener({ point: { x: 10, y: 10 } });
      }
    }

    getLayer() {
      return {};
    }

    getSource() {
      return {};
    }

    addControl() {}

    setMinZoom() {}

    fitBounds() {}

    getCanvas() {
      return { style: {} };
    }

    getContainer() {
      return document.createElement("div");
    }

    remove() {}
  }

  const created: FakeMap[] = [];
  return { FakeMap, created };
});

vi.mock("maplibre-gl", () => ({
  Map: maps.FakeMap,
  AttributionControl: class {},
  NavigationControl: class {},
  setWorkerUrl: () => {},
}));

afterEach(() => {
  maps.created.length = 0;
});

function renderMap(onSelect = vi.fn()) {
  render(
    <MapLibreMap
      locations={[]}
      mode="all"
      backend={{ kind: "outline" }}
      onSelect={onSelect}
    />,
  );

  return { onSelect, map: maps.created[0] };
}

describe("MapLibreMap", () => {
  it("clears the selection on a click away from every marker", () => {
    const { onSelect, map } = renderMap();

    map.clickAt([]);

    expect(onSelect).toHaveBeenCalledWith(null);
  });

  it("selects the area of the marker clicked", async () => {
    const { onSelect, map } = renderMap();

    map.clickAt([
      {
        properties: { devices: 2, online: 1, longitude: 8.54, latitude: 47.38 },
        geometry: { type: "Point", coordinates: [8.54, 47.38] },
      },
    ]);

    await waitFor(() =>
      expect(onSelect).toHaveBeenCalledWith(
        expect.objectContaining({ devices: 2, online: 1 }),
      ),
    );
    const [[{ bounds }]] = onSelect.mock.calls as [[{ bounds: number[] }]];
    const [west, south, east, north] = bounds;
    expect(west).toBeLessThan(8.54);
    expect(east).toBeGreaterThan(8.54);
    expect(south).toBeLessThan(47.38);
    expect(north).toBeGreaterThan(47.38);
    expect(onSelect).not.toHaveBeenCalledWith(null);
  });
});
