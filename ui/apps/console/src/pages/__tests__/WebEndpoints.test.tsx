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
  it("shows the endpoint count when totalCount > 0 and only one page exists", async () => {
    setEndpoints([ep("ep1.example.com")], 1);
    renderPage();
    expect(await screen.findByText(/1 endpoint/i)).toBeInTheDocument();
  });

  it("hides Prev/Next controls when only one page exists", async () => {
    setEndpoints([ep("ep1.example.com")], 1);
    renderPage();
    await screen.findByText(/1 endpoint/i);
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
    server.use(
      http.get("*/api/web-endpoints", () => new Promise(() => {})),
    );
    renderPage();
    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });
});

describe("WebEndpoints — URL hydration", () => {
  it("passes page=2 and a filter containing the search term from the URL", async () => {
    setEndpoints([ep("ep1.example.com")], 1);
    renderPage(["/?page=2&search=myhost"]);
    await waitFor(() => {
      expect(lastRequestUrl).not.toBeNull();
      expect(lastRequestUrl!.searchParams.get("page")).toBe("2");
      const filter = lastRequestUrl!.searchParams.get("filter") ?? "";
      expect(atob(filter)).toContain("myhost");
    });
  });

  it("falls back to page=1 and no filter when URL has no params", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      expect(lastRequestUrl).not.toBeNull();
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
      expect(lastRequestUrl!.searchParams.get("filter")).toBeNull();
    });
  });
});

describe("WebEndpoints — search resets page to 1", () => {
  it("resets to page=1 when a new search term is typed while on page 2", async () => {
    const user = userEvent.setup();
    setEndpoints([ep("ep1.example.com")], 25);
    renderPage(["/?page=2"]);
    await screen.findByText(/25 endpoints/i);

    const searchInput = screen.getByPlaceholderText(/search by address/i);
    await user.type(searchInput, "x");

    await waitFor(() => {
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
    });
  });
});

describe("WebEndpoints — page change writes to URL", () => {
  it("calls listWebEndpoints with page=2 after clicking the Next page button", async () => {
    const user = userEvent.setup();
    setEndpoints(
      Array.from({ length: 10 }, (_, i) => ep(`ep${i + 1}.example.com`)),
      15,
    );
    renderPage(["/"]);
    await screen.findByText(/15 endpoints/i);

    await user.click(screen.getByRole("button", { name: /next page/i }));

    await waitFor(() => {
      expect(lastRequestUrl!.searchParams.get("page")).toBe("2");
    });
  });
});

async function openEndpointDrawer(user: ReturnType<typeof userEvent.setup>) {
  setEndpoints([ep("ep1.example.com")], 1);
  renderPage();
  await user.click(
    await screen.findByRole("button", { name: /new endpoint/i }),
  );
}

describe("WebEndpoints — expiration toggle", () => {
  it("exposes the expiration control as a switch with aria-checked and no aria-pressed", async () => {
    const user = userEvent.setup();
    await openEndpointDrawer(user);
    const toggle = screen.getByRole("switch", { name: /set expiration/i });
    expect(toggle).toHaveAttribute("aria-checked", "false");
    expect(toggle).not.toHaveAttribute("aria-pressed");
  });

  it("flips aria-checked when the expiration switch is clicked", async () => {
    const user = userEvent.setup();
    await openEndpointDrawer(user);
    const toggle = screen.getByRole("switch", { name: /set expiration/i });
    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-checked", "true");
  });
});

describe("WebEndpoints — TLS toggle", () => {
  it("exposes the TLS control as a switch with aria-checked reflecting state", async () => {
    const user = userEvent.setup();
    await openEndpointDrawer(user);
    const toggle = screen.getByRole("switch", { name: /uses https/i });
    expect(toggle).toHaveAttribute("aria-checked", "false");
  });

  it("calls the TLS handler with the toggled boolean when clicked", async () => {
    const user = userEvent.setup();
    await openEndpointDrawer(user);
    const toggle = screen.getByRole("switch", { name: /uses https/i });
    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-checked", "true");
  });
});
