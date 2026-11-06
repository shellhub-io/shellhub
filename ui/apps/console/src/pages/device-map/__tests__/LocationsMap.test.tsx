import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import LocationsMap from "../LocationsMap";

vi.mock("@/pages/device-map/MapLibreMap", () => ({
  default: () => <div data-testid="maplibre-map" />,
}));

function renderMap() {
  return render(<LocationsMap locations={[]} mode="all" source="dbip" />);
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("LocationsMap", () => {
  it("says the browser cannot draw the map when it has no WebGL 2", async () => {
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(null);

    renderMap();

    expect(await screen.findByText(/this browser cannot draw the map/i)).toBeInTheDocument();
    expect(screen.queryByTestId("maplibre-map")).not.toBeInTheDocument();
  });

  it("draws the map when the browser has WebGL 2", async () => {
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(
      {} as WebGL2RenderingContext,
    );

    renderMap();

    expect(await screen.findByTestId("maplibre-map")).toBeInTheDocument();
    expect(screen.queryByText(/this browser cannot draw the map/i)).not.toBeInTheDocument();
  });
});
