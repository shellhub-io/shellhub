import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { CurrentUrl } from "@/tests/CurrentUrl";
import { mockDevice, mockStats } from "@/tests/factories";
import { defaultHandlers } from "@/tests/handlers";
import DeviceMap from "../DeviceMap";
import type { LocationsMapProps } from "../mapModel";

const ZURICH_AREA = "8.5349,47.3749,8.5451,47.3851";

vi.mock("@/pages/device-map/MapLibreMap", () => ({
  default: ({ onSelect }: LocationsMapProps) => (
    <button
      type="button"
      onClick={() =>
        onSelect?.({
          bounds: [8.5349, 47.3749, 8.5451, 47.3851],
          devices: 0,
          online: 0,
        })
      }
    >
      Pick Zurich
    </button>
  ),
}));

function serveMap({ located, listed }: { located: number; listed: number }) {
  const requests: URL[] = [];
  server.use(...defaultHandlers);
  server.use(
    http.get("*/api/devices/locations", () =>
      HttpResponse.json(
        [{ latitude: 47.38, longitude: 8.54, devices: located, online: located }],
        { headers: { "X-GeoIP-Source": "dbip" } },
      ),
    ),
    http.get("*/api/devices", ({ request }) => {
      requests.push(new URL(request.url));

      return jsonWithTotal([mockDevice()], listed);
    }),
    http.get("*/api/stats", () =>
      HttpResponse.json(mockStats({ registered_devices: listed + 20 })),
    ),
  );

  return requests;
}

function renderMap(url = "/dashboard/map") {
  return render(
    <>
      <DeviceMap />
      <CurrentUrl />
    </>,
    { wrapper: createTestWrapper({ initialEntries: [url] }) },
  );
}

beforeEach(() => {
  seedAuthStore();
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(
    {} as WebGL2RenderingContext,
  );
});

describe("DeviceMap", () => {
  it("counts as having no position only the listed devices the map does not place", async () => {
    const requests = serveMap({ located: 10, listed: 13 });

    renderMap();

    const noPosition = (await screen.findByText("No position")).parentElement;
    if (!noPosition) throw new Error("no figure holds the No position label");
    expect(await within(noPosition).findByText("3")).toBeInTheDocument();
    expect(screen.queryByText("23")).not.toBeInTheDocument();
    expect(requests.length).toBeGreaterThan(0);
    expect(requests.every((url) => url.searchParams.get("status") === "accepted")).toBe(true);
    expect(requests.every((url) => url.searchParams.get("connector") !== "true")).toBe(true);
  });

  it("opens the area a link names, counting the devices the map places in it", async () => {
    serveMap({ located: 10, listed: 13 });

    renderMap(`/dashboard/map?area=${ZURICH_AREA}`);

    const panel = await screen.findByRole("complementary");
    expect(within(panel).getByText("This area")).toBeInTheDocument();
    expect(within(panel).getByText("10 devices")).toBeInTheDocument();
  });

  it("opens the brief of the one device at a point a link names", async () => {
    serveMap({ located: 1, listed: 1 });

    renderMap(`/dashboard/map?area=${ZURICH_AREA}`);

    expect(
      await screen.findByRole("dialog", { name: mockDevice().name }),
    ).toBeInTheDocument();
  });

  it("opens no area while the locations cannot be read", async () => {
    serveMap({ located: 10, listed: 13 });
    server.use(
      http.get("*/api/devices/locations", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );

    renderMap(`/dashboard/map?area=${ZURICH_AREA}`);

    expect(
      await screen.findByText("Could not load device locations."),
    ).toBeInTheDocument();
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });

  it("ignores an area that does not parse", async () => {
    serveMap({ located: 10, listed: 13 });

    renderMap("/dashboard/map?area=nowhere");

    const located = screen.getByText("Located").parentElement;
    if (!located) throw new Error("no figure holds the Located label");
    expect(await within(located).findByText("10")).toBeInTheDocument();
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });

  it("names the picked area in the URL and drops it when the panel closes", async () => {
    const user = userEvent.setup();
    serveMap({ located: 10, listed: 13 });

    renderMap();
    await user.click(
      await screen.findByRole("button", { name: "Pick Zurich" }),
    );

    expect(screen.getByRole("status", { name: "URL" })).toHaveTextContent(
      `/dashboard/map?area=${ZURICH_AREA}`,
    );
    expect(await screen.findByText("10 devices")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Close area" }));

    expect(screen.getByRole("status", { name: "URL" })).toHaveTextContent(
      /^\/dashboard\/map$/,
    );
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });
});
