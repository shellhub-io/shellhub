import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
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

const drawnMap = vi.hoisted(() => ({
  props: undefined as LocationsMapProps | undefined,
}));

vi.mock("@/pages/device-map/MapLibreMap", () => ({
  default: (props: LocationsMapProps) => {
    drawnMap.props = props;

    return (
      <>
        <button
          type="button"
          onClick={() =>
            props.onSelect?.({
              bounds: [8.5349, 47.3749, 8.5451, 47.3851],
              devices: 0,
              online: 0,
            })
          }
        >
          Pick Zurich
        </button>
        <button type="button" onClick={() => props.onSelect?.(null)}>
          Click away
        </button>
      </>
    );
  },
}));

function serveMap({ located, listed }: { located: number; listed: number }) {
  const requests: URL[] = [];
  server.use(...defaultHandlers);
  server.use(
    http.get("*/api/devices/locations", () =>
      HttpResponse.json(
        [
          {
            latitude: 47.38,
            longitude: 8.54,
            devices: located,
            online: located,
          },
        ],
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

async function pickZurich(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: "Pick Zurich" }));
  await screen.findByText("10 devices");
}

function paddingGivenToMap() {
  const padding = drawnMap.props?.areaPadding;
  if (!padding) throw new Error("the map was drawn with no area padding");

  return padding;
}

beforeEach(() => {
  drawnMap.props = undefined;
  localStorage.clear();
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
    expect(
      requests.every((url) => url.searchParams.get("status") === "accepted"),
    ).toBe(true);
    expect(
      requests.every((url) => url.searchParams.get("connector") !== "true"),
    ).toBe(true);
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

  it("drops the picked area on a click away from every marker", async () => {
    const user = userEvent.setup();
    serveMap({ located: 10, listed: 13 });

    renderMap(`/dashboard/map?area=${ZURICH_AREA}`);
    expect(await screen.findByText("10 devices")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Click away" }));

    expect(screen.getByRole("status", { name: "URL" })).toHaveTextContent(
      /^\/dashboard\/map$/,
    );
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });

  describe("folding", () => {
    it("shows the online count in a folded summary's title row", async () => {
      const user = userEvent.setup();
      serveMap({ located: 10, listed: 13 });
      renderMap();

      await user.click(screen.getByRole("button", { name: "Map details" }));

      const titleRow = screen.getByRole("heading", {
        name: "Device Map",
      }).parentElement;
      await waitFor(() => expect(titleRow).toHaveTextContent("10/10 online"));
    });

    it("keeps the mode switch of a folded summary", async () => {
      const user = userEvent.setup();
      serveMap({ located: 10, listed: 13 });
      renderMap();

      await user.click(screen.getByRole("button", { name: "Map details" }));

      expect(screen.getByText(/GeoIP lookup/)).not.toBeVisible();
      expect(
        screen.getByRole("radiogroup", { name: "Devices shown" }),
      ).toBeVisible();
    });

    it("keeps the summary folded across a reload", async () => {
      const user = userEvent.setup();
      serveMap({ located: 10, listed: 13 });
      const { unmount } = renderMap();

      await user.click(screen.getByRole("button", { name: "Map details" }));
      unmount();
      renderMap();

      expect(
        screen.getByRole("button", { name: "Map details" }),
      ).toHaveAttribute("aria-expanded", "false");
      expect(screen.getByText(/GeoIP lookup/)).not.toBeVisible();
    });

    it("keeps the area panel folded when it closes and an area is picked again", async () => {
      const user = userEvent.setup();
      serveMap({ located: 10, listed: 13 });
      renderMap();
      await pickZurich(user);

      await user.click(
        screen.getByRole("button", { name: "Devices in this area" }),
      );
      await user.click(screen.getByRole("button", { name: "Close area" }));
      await pickZurich(user);

      expect(
        screen.getByRole("button", { name: "Devices in this area" }),
      ).toHaveAttribute("aria-expanded", "false");
    });

    it("hands the map narrower side padding once both panels fold", async () => {
      const user = userEvent.setup();
      serveMap({ located: 10, listed: 13 });
      renderMap();
      await pickZurich(user);
      const open = paddingGivenToMap();

      await user.click(
        screen.getByRole("button", { name: "Devices in this area" }),
      );
      await user.click(screen.getByRole("button", { name: "Map details" }));

      expect(paddingGivenToMap().right).toBeLessThan(open.right);
      expect(paddingGivenToMap().left).toBeLessThan(open.left);
    });
  });
});
