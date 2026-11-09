import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { CurrentUrl } from "@/tests/CurrentUrl";
import { mockNamespace, mockStats } from "@/tests/factories";
import type { DeviceLocation } from "@/client";
import type { LocationsMapProps } from "@/pages/device-map/mapModel";
import Dashboard from "../Dashboard";

vi.mock("@/pages/device-map/MapLibreMap", () => ({
  default: ({ onSelect }: LocationsMapProps) =>
    onSelect ? (
      <>
        <button
          type="button"
          onClick={() =>
            onSelect({
              bounds: [8.534899999, 47.3749, 8.5451, 47.3851],
              devices: 9,
              online: 8,
            })
          }
        >
          Pick Zurich
        </button>
        <button type="button" onClick={() => onSelect(null)}>
          Click away
        </button>
      </>
    ) : null,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

const zurich: DeviceLocation = {
  latitude: 47.38,
  longitude: 8.54,
  devices: 9,
  online: 8,
};

function serveDashboard({
  locations = [zurich],
  locationsHeld,
  locationsFail = false,
  geoipSource = "dbip",
}: {
  locations?: DeviceLocation[];
  locationsHeld?: Promise<void>;
  locationsFail?: boolean;
  geoipSource?: string;
} = {}) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/stats", () =>
      HttpResponse.json(
        mockStats({ registered_devices: 10, online_devices: 9 }),
      ),
    ),
    http.get("*/api/sessions", () => jsonWithTotal([])),
    http.get("*/api/devices/locations", async () => {
      await locationsHeld;
      if (locationsFail) return HttpResponse.json({}, { status: 500 });

      return HttpResponse.json(locations, {
        headers: geoipSource ? { "X-GeoIP-Source": geoipSource } : {},
      });
    }),
  );
}

function renderPage(queryClient = createTestQueryClient()) {
  return render(
    <>
      <Dashboard />
      <CurrentUrl />
    </>,
    {
      wrapper: createTestWrapper({
        initialEntries: ["/dashboard"],
        queryClient,
      }),
    },
  );
}

beforeEach(() => {
  seedAuthStore();
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(
    {} as WebGL2RenderingContext,
  );
});

describe("Dashboard", () => {
  it("shows the online and offline devices with their share of the accepted ones", async () => {
    serveDashboard();

    renderPage();

    expect(
      await screen.findByRole("link", { name: /online\s*9\s*90%/i }),
    ).toHaveAttribute("href", "/devices");
    expect(
      screen.getByRole("link", { name: /offline\s*1\s*10%/i }),
    ).toHaveAttribute("href", "/devices");
  });

  it("opens the full map through one link once a device has a position", async () => {
    serveDashboard();

    renderPage();

    expect(
      await screen.findByRole("link", { name: /view full map/i }),
    ).toHaveAttribute("href", "/dashboard/map");
    expect(
      screen
        .getAllByRole("link")
        .filter((link) => link.getAttribute("href") === "/dashboard/map"),
    ).toHaveLength(1);
  });

  it("opens the full map on the area picked on the map", async () => {
    const user = userEvent.setup();
    serveDashboard();

    renderPage();
    await user.click(
      await screen.findByRole("button", { name: "Pick Zurich" }),
    );

    expect(screen.getByRole("status", { name: "URL" })).toHaveTextContent(
      "/dashboard/map?area=8.5349,47.3749,8.5451,47.3851",
    );
  });

  it("stays on the Dashboard on a click away from every marker", async () => {
    const user = userEvent.setup();
    serveDashboard();

    renderPage();
    await user.click(await screen.findByRole("button", { name: "Click away" }));

    expect(screen.getByRole("status", { name: "URL" })).toHaveTextContent(
      /^\/dashboard$/,
    );
  });

  it("says no device has a position and offers no full map while none has", async () => {
    serveDashboard({ locations: [] });

    renderPage();

    expect(
      await screen.findByText(/no device has a position yet/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /full device map|view full map/i }),
    ).not.toBeInTheDocument();
  });

  it("says the positions could not be loaded when reading them fails", async () => {
    serveDashboard({ locationsFail: true });

    renderPage();

    expect(
      await screen.findByText(/could not load device positions/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/no device has a position yet/i),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /view full map/i }),
    ).not.toBeInTheDocument();
  });

  it("offers no full map once a refresh of the positions fails", async () => {
    serveDashboard();
    const queryClient = createTestQueryClient();
    renderPage(queryClient);
    await screen.findByRole("link", { name: /view full map/i });

    serveDashboard({ locationsFail: true });
    await act(() =>
      queryClient.refetchQueries({ queryKey: [{ _id: "getDeviceLocations" }] }),
    );

    expect(
      await screen.findByText(/could not load device positions/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /view full map/i }),
    ).not.toBeInTheDocument();
  });

  it("shows the figures without waiting for the device positions", async () => {
    let releaseLocations = () => {};
    const locationsHeld = new Promise<void>((resolve) => {
      releaseLocations = resolve;
    });
    serveDashboard({ locationsHeld });

    renderPage();

    expect(
      await screen.findByRole("link", { name: /online\s*9\s*90%/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/no device has a position yet/i),
    ).not.toBeInTheDocument();

    releaseLocations();
    await screen.findByRole("link", { name: /view full map/i });
  });

  it("credits DB-IP when the server places the devices with it", async () => {
    serveDashboard({ geoipSource: "dbip" });

    renderPage();

    expect(
      await screen.findByRole("link", { name: /ip geolocation by db-ip/i }),
    ).toHaveAttribute("href", "https://db-ip.com");
  });

  it("credits MaxMind when the server places the devices with GeoLite2", async () => {
    serveDashboard({ geoipSource: "maxmind" });

    renderPage();

    expect(
      await screen.findByText(/geolite data created by maxmind/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /db-ip/i }),
    ).not.toBeInTheDocument();
  });

  it("credits no database when the server names none", async () => {
    serveDashboard({ geoipSource: "" });

    renderPage();

    await screen.findByRole("link", { name: /online\s*9\s*90%/i });
    expect(screen.queryByText(/db-ip|maxmind/i)).not.toBeInTheDocument();
  });
});
