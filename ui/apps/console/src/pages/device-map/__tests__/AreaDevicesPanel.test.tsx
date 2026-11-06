import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice } from "@/tests/factories";
import { decodeB64url } from "@/tests/decodeB64url";
import AreaDevicesPanel from "../AreaDevicesPanel";
import type { MapArea } from "../mapModel";

const area: MapArea = {
  bounds: [8.5, 47.3, 8.6, 47.4],
  devices: 3,
  online: 2,
};

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

function renderPanel() {
  return render(<AreaDevicesPanel area={area} onClose={() => {}} />, {
    wrapper: createTestWrapper({ initialEntries: ["/dashboard/map"] }),
  });
}

beforeEach(() => {
  seedAuthStore();
  observers.clear();
  vi.stubGlobal("IntersectionObserver", ReachableObserver);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("AreaDevicesPanel", () => {
  it("lists the first page of the devices inside the area", async () => {
    const requests = servePages();

    renderPanel();

    expect(
      await screen.findByRole("link", { name: /zurich-1/ }),
    ).toHaveAttribute("href", "/devices/a");
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

  it("reads the next page by the cursor the last one handed back when the end is reached", async () => {
    const requests = servePages();
    renderPanel();
    await screen.findByRole("link", { name: /zurich-2/ });

    reachEnd();

    expect(
      await screen.findByRole("link", { name: /zurich-3/ }),
    ).toBeInTheDocument();
    expect(requests[1].searchParams.get("page[after]")).toBe("cursor-1");
    expect(screen.getByText("3 loaded")).toBeInTheDocument();
  });

  it("reads nothing more once a page comes back without a cursor", async () => {
    const requests = servePages();
    renderPanel();
    await screen.findByRole("link", { name: /zurich-2/ });
    reachEnd();
    await screen.findByRole("link", { name: /zurich-3/ });

    reachEnd();

    expect(requests).toHaveLength(2);
  });
});
