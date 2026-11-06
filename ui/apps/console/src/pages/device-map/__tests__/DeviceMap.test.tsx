import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice, mockStats } from "@/tests/factories";
import DeviceMap from "../DeviceMap";

vi.mock("@/pages/device-map/MapLibreMap", () => ({
  default: () => null,
}));

function serveMap({ located, listed }: { located: number; listed: number }) {
  const requests: URL[] = [];
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

function renderMap() {
  return render(<DeviceMap />, {
    wrapper: createTestWrapper({ initialEntries: ["/dashboard/map"] }),
  });
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
});
