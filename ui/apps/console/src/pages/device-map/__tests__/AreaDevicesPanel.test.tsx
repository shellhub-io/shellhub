import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  onTestFinished,
  vi,
} from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { onlineManager, type QueryClient } from "@tanstack/react-query";
import { server } from "@/tests/msw";
import type { Device } from "@/client";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { mockDevice } from "@/tests/factories";
import { defaultHandlers } from "@/tests/handlers";
import { decodeB64url } from "@/tests/decodeB64url";
import AreaDevicesPanel from "../AreaDevicesPanel";
import type { MapArea } from "../mapModel";

const area: MapArea = {
  bounds: [8.5, 47.3, 8.6, 47.4],
  devices: 3,
  online: 2,
};

const point: MapArea = { ...area, devices: 1, online: 1 };

const observers = new Set<{
  callback: IntersectionObserverCallback;
  active: boolean;
}>();

class ReachableObserver {
  private entry: { callback: IntersectionObserverCallback; active: boolean };

  constructor(callback: IntersectionObserverCallback) {
    this.entry = { callback, active: false };
    observers.add(this.entry);
  }

  observe() {
    this.entry.active = true;
  }

  disconnect() {
    this.entry.active = false;
  }
}

function reachEnd() {
  act(() => {
    for (const observer of observers) {
      if (!observer.active) continue;
      observer.callback(
        [{ isIntersecting: true } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
    }
  });
}

function servePages() {
  const requests: URL[] = [];
  server.use(
    http.get("*/api/devices", ({ request }) => {
      const url = new URL(request.url);
      requests.push(url);

      if (!url.searchParams.has("page[after]")) {
        return HttpResponse.json(
          [
            mockDevice({ uid: "a", name: "zurich-1" }),
            mockDevice({ uid: "b", name: "zurich-2" }),
          ],
          { headers: { "X-Next-Cursor": "cursor-1" } },
        );
      }

      return HttpResponse.json([mockDevice({ uid: "c", name: "zurich-3" })]);
    }),
  );

  return requests;
}

function serveDevices(devices: Device[]) {
  const requests: URL[] = [];
  server.use(
    ...defaultHandlers,
    http.get("*/api/devices", ({ request }) => {
      requests.push(new URL(request.url));

      return HttpResponse.json(devices);
    }),
  );

  return requests;
}

function Panel({
  area: picked,
  folded = false,
  onClose = () => {},
}: {
  area: MapArea;
  folded?: boolean;
  onClose?: () => void;
}) {
  return (
    <AreaDevicesPanel
      area={picked}
      folded={folded}
      onToggleFold={() => {}}
      onClose={onClose}
    />
  );
}

function renderPanel({
  picked = area,
  queryClient = createTestQueryClient(),
  onClose = () => {},
}: {
  picked?: MapArea;
  queryClient?: QueryClient;
  onClose?: () => void;
} = {}) {
  return render(<Panel area={picked} onClose={onClose} />, {
    wrapper: createTestWrapper({
      queryClient,
      initialEntries: ["/dashboard/map"],
    }),
  });
}

beforeEach(() => {
  seedAuthStore();
  observers.clear();
  vi.stubGlobal("IntersectionObserver", ReachableObserver);
});

afterEach(() => {
  vi.unstubAllGlobals();
  onlineManager.setOnline(true);
});

describe("AreaDevicesPanel", () => {
  it("lists the first page of the devices inside the area", async () => {
    const requests = servePages();

    renderPanel();

    expect(
      await screen.findByRole("button", { name: /zurich-1/ }),
    ).toBeInTheDocument();
    const filter = decodeB64url(
      requests[0].searchParams.get("filter") ?? "",
    ) as { params: { name?: string; operator?: string; value?: number } }[];
    expect(filter.map(({ params }) => params)).toEqual([
      { name: "and" },
      { name: "latitude", operator: "gt", value: 47.3 },
      { name: "latitude", operator: "lt", value: 47.4 },
      { name: "longitude", operator: "gt", value: 8.5 },
      { name: "longitude", operator: "lt", value: 8.6 },
      { name: "located", operator: "bool", value: true },
    ]);
  });

  it("opens a brief of the device picked in the list", async () => {
    const user = userEvent.setup();
    servePages();
    server.use(...defaultHandlers);
    renderPanel();

    await user.click(await screen.findByRole("button", { name: /zurich-2/ }));

    expect(
      screen.getByRole("dialog", { name: "zurich-2" }),
    ).toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    renderPanel({ onClose });

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("stays open on Escape while a terminal window covers the map", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    const sessions = useTerminalStore.getState().sessions;
    useTerminalStore.setState({
      sessions: [
        {
          id: "session-1",
          deviceUid: "device-uid",
          deviceName: "my-device",
          username: "root",
          password: "",
          state: "shown",
          connectionStatus: "connected",
        },
      ],
    });
    onTestFinished(() => useTerminalStore.setState({ sessions }));
    renderPanel({ onClose });

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  it("stays open on Escape while a brief is open over it", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    server.use(...defaultHandlers);
    renderPanel({ onClose });
    await user.click(await screen.findByRole("button", { name: /zurich-2/ }));

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  function mountBeside(attributes: Record<string, string>, tag = "div") {
    const element = document.body.appendChild(document.createElement(tag));
    for (const [name, value] of Object.entries(attributes)) {
      element.setAttribute(name, value);
    }
    onTestFinished(() => element.remove());
  }

  it.each([
    ["an open native dialog", "dialog", { open: "" }],
    ["an open modal drawer", "div", { role: "dialog", "aria-modal": "true" }],
  ])("stays open on Escape while %s is over it", async (_, tag, attributes) => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    renderPanel({ onClose });
    mountBeside(attributes, tag);

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  it("stays open on an Escape that closes a modal drawer on its way up", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    renderPanel({ onClose });
    const drawer = document.body.appendChild(document.createElement("div"));
    drawer.setAttribute("role", "dialog");
    drawer.setAttribute("aria-modal", "true");
    drawer.tabIndex = -1;
    drawer.addEventListener("keydown", () => {
      drawer.setAttribute("aria-modal", "false");
      drawer.setAttribute("inert", "");
    });
    onTestFinished(() => drawer.remove());
    drawer.focus();

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  it("stays open on an Escape that a menu stops on its way up", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    renderPanel({ onClose });
    const menu = document.body.appendChild(document.createElement("div"));
    menu.setAttribute("role", "menu");
    menu.tabIndex = -1;
    menu.addEventListener("keydown", (event) => event.stopPropagation());
    onTestFinished(() => menu.remove());
    menu.focus();

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  it("closes on Escape while a modal drawer sits closed and inert", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    servePages();
    renderPanel({ onClose });
    mountBeside({ role: "dialog", "aria-modal": "true", inert: "" });

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("reads the next page by the cursor the last one handed back when the end is reached", async () => {
    const requests = servePages();
    renderPanel();
    await screen.findByRole("button", { name: /zurich-2/ });

    reachEnd();

    expect(
      await screen.findByRole("button", { name: /zurich-3/ }),
    ).toBeInTheDocument();
    expect(requests[1].searchParams.get("page[after]")).toBe("cursor-1");
    expect(screen.getByText("3 loaded")).toBeInTheDocument();
  });

  it("reads nothing more once a page comes back without a cursor", async () => {
    const requests = servePages();
    renderPanel();
    await screen.findByRole("button", { name: /zurich-2/ });
    reachEnd();
    await screen.findByRole("button", { name: /zurich-3/ });

    reachEnd();

    expect(requests).toHaveLength(2);
  });

  it("opens the brief of the one device a picked point holds", async () => {
    serveDevices([mockDevice({ name: "zurich-1" })]);

    renderPanel({ picked: point });

    expect(
      await screen.findByRole("dialog", { name: "zurich-1" }),
    ).toBeInTheDocument();
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });

  it("reads a picked point's device among the accepted ones", async () => {
    const requests = serveDevices([mockDevice()]);

    renderPanel({ picked: point });

    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0].searchParams.get("status")).toBe("accepted");
  });

  it("keeps a picked point's brief open when the device is read again", async () => {
    const queryClient = createTestQueryClient();
    serveDevices([mockDevice({ name: "zurich-1" })]);
    renderPanel({ picked: point, queryClient });
    await screen.findByRole("dialog", { name: "zurich-1" });
    serveDevices([
      mockDevice({ uid: "a", name: "zurich-1" }),
      mockDevice({ uid: "b", name: "zurich-2" }),
    ]);

    await act(() => queryClient.refetchQueries());

    expect(
      screen.getByRole("dialog", { name: "zurich-1" }),
    ).toBeInTheDocument();
    expect(screen.queryByText("This area")).not.toBeInTheDocument();
  });

  it("lists the devices of a picked point when its bounds hold more than one", async () => {
    serveDevices([
      mockDevice({ uid: "a", name: "zurich-1" }),
      mockDevice({ uid: "b", name: "zurich-2" }),
    ]);

    renderPanel({ picked: point });

    expect(
      await screen.findByRole("button", { name: /zurich-2/ }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows only the header of a folded area and keeps its list for unfolding", async () => {
    servePages();
    const { rerender } = renderPanel();
    const device = await screen.findByRole("button", { name: /zurich-1/ });

    rerender(<Panel area={area} folded />);

    expect(screen.getByText("3 devices")).toBeVisible();
    expect(device).toBeInTheDocument();
    expect(device).not.toBeVisible();
    expect(
      screen.getByRole("button", { name: "Devices in this area" }),
    ).toHaveAttribute("aria-expanded", "false");
  });

  it("keeps the list of an area whose count falls to one while it is open", async () => {
    serveDevices([mockDevice({ uid: "a", name: "zurich-1" })]);
    const { rerender } = renderPanel({
      picked: { ...area, devices: 2, online: 2 },
    });
    await screen.findByRole("button", { name: /zurich-1/ });

    rerender(<Panel area={point} />);

    expect(
      screen.getByRole("button", { name: /zurich-1/ }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Loading device")).not.toBeInTheDocument();
  });

  it("reads a picked point's devices once when it falls back to the list", async () => {
    const requests = serveDevices([
      mockDevice({ uid: "a", name: "zurich-1" }),
      mockDevice({ uid: "b", name: "zurich-2" }),
    ]);

    renderPanel({ picked: point });

    await screen.findByRole("button", { name: /zurich-2/ });
    expect(requests).toHaveLength(1);
  });

  it("lists the devices of a picked point when its bounds hold none", async () => {
    serveDevices([]);

    renderPanel({ picked: point });

    expect(await screen.findByText("This area")).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("lists the devices of a picked point when its device cannot be read", async () => {
    server.use(
      http.get("*/api/devices", () => HttpResponse.json({}, { status: 500 })),
    );

    renderPanel({ picked: point });

    expect(
      await screen.findByText("Could not load the devices in this area."),
    ).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("waits for the network to read a picked point's device", async () => {
    serveDevices([mockDevice({ name: "zurich-1" })]);
    onlineManager.setOnline(false);
    renderPanel({ picked: point });
    expect(
      await screen.findByRole("status", { name: "Loading device" }),
    ).toBeInTheDocument();

    act(() => onlineManager.setOnline(true));

    expect(
      await screen.findByRole("dialog", { name: "zurich-1" }),
    ).toBeInTheDocument();
  });

  it("shows it is loading the picked point's device", async () => {
    server.use(http.get("*/api/devices", () => new Promise(() => {})));

    renderPanel({ picked: point });

    expect(
      await screen.findByRole("status", { name: "Loading device" }),
    ).toBeInTheDocument();
  });
});
