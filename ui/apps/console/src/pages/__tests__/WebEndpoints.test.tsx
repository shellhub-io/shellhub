import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockWebEndpoint } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import WebEndpoints from "../WebEndpoints";

vi.mock("@/hooks/useResetOnOpen");

vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: vi.fn(<T,>(value: T) => value),
}));

import { useDebouncedValue } from "@/hooks/useDebouncedValue";

const mockUseDebouncedValue = vi.mocked(useDebouncedValue);

function ep(address: string) {
  return mockWebEndpoint({ address, full_address: address });
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(<WebEndpoints />, {
    wrapper: createTestWrapper({ initialEntries }),
  });
}

let lastRequestUrl: URL | null;

function setEndpoints(endpoints: ReturnType<typeof ep>[], total?: number) {
  server.use(
    http.get("*/api/web-endpoints", ({ request }) => {
      lastRequestUrl = new URL(request.url);
      return jsonWithTotal(endpoints, total ?? endpoints.length);
    }),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  lastRequestUrl = null;
  seedAuthStore();
  setEndpoints([]);
  server.use(
    http.delete(
      "*/api/web-endpoints/:address",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post("*/api/web-endpoints", () =>
      HttpResponse.json(ep("new.example.com")),
    ),
    http.get("*/api/devices", () => jsonWithTotal([])),
  );
  mockUseDebouncedValue.mockImplementation(<T,>(v: T) => v);
});

describe("WebEndpoints — pagination count / controls decoupling", () => {
  it("shows the count but no Prev/Next controls when only one page exists", async () => {
    setEndpoints([ep("ep1.example.com")], 1);
    renderPage();
    expect(await screen.findByText(/1 endpoint/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /next page/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Prev/Next controls and the count when multiple pages exist", async () => {
    const endpoints = Array.from({ length: 10 }, (_, i) =>
      ep(`ep${i + 1}.example.com`),
    );
    setEndpoints(endpoints, 15);
    renderPage();
    expect(await screen.findByText(/15 endpoints/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /previous page/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /next page/i }),
    ).toBeInTheDocument();
  });

  it("does not show the Pagination nav when there are no endpoints", async () => {
    renderPage();
    await waitFor(() => expect(lastRequestUrl).not.toBeNull());
    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });

  it("does not flash a '0 endpoints' count while a search request is in-flight", () => {
    mockUseDebouncedValue.mockReturnValue("some-query");
    server.use(http.get("*/api/web-endpoints", () => new Promise(() => {})));
    renderPage();
    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });
});

async function openEndpointDrawer(user: ReturnType<typeof userEvent.setup>) {
  setEndpoints([ep("ep1.example.com")], 1);
  renderPage();
  await user.click(
    await screen.findByRole("button", { name: /new endpoint/i }),
  );
}

describe("WebEndpoints — drawer toggles", () => {
  it.each([/set expiration/i, /uses https/i])(
    "exposes %s as a switch whose aria-checked flips on click",
    async (name) => {
      const user = userEvent.setup();
      await openEndpointDrawer(user);
      const toggle = screen.getByRole("switch", { name });
      expect(toggle).toHaveAttribute("aria-checked", "false");
      expect(toggle).not.toHaveAttribute("aria-pressed");

      await user.click(toggle);

      expect(toggle).toHaveAttribute("aria-checked", "true");
    },
  );
});
