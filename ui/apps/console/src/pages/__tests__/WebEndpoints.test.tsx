import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { Webendpoint } from "@/client/types.gen";
import WebEndpoints from "../WebEndpoints";

vi.mock("@/hooks/useWebEndpoints");
vi.mock("@/hooks/useWebEndpointMutations");
vi.mock("@/hooks/useDevices");
vi.mock("@/hooks/useResetOnOpen");
vi.mock("@/hooks/useHasPermission");
vi.mock("@/hooks/useDebouncedValue");

import { useWebEndpoints } from "@/hooks/useWebEndpoints";
import {
  useDeleteWebEndpoint,
  useCreateWebEndpoint,
} from "@/hooks/useWebEndpointMutations";
import { useDevices } from "@/hooks/useDevices";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";

const mockUseWebEndpoints = vi.mocked(useWebEndpoints);
const mockUseDeleteWebEndpoint = vi.mocked(useDeleteWebEndpoint);
const mockUseCreateWebEndpoint = vi.mocked(useCreateWebEndpoint);
const mockUseDevices = vi.mocked(useDevices);
const mockUseHasPermission = vi.mocked(useHasPermission);
const mockUseDebouncedValue = vi.mocked(useDebouncedValue);

function makeEndpoint(address: string): Webendpoint {
  return {
    address,
    full_address: address,
    namespace: "00000000-0000-4000-0000-000000000000",
    device_uid: "dev-uid",
    host: "localhost",
    port: 8080,
    ttl: 30,
    expires_in: "0001-01-01T00:00:00Z",
    created_at: "2024-01-15T10:00:00Z",
    device: { name: "my-device", uid: "dev-uid" },
  } as unknown as Webendpoint;
}

function setupDefaultMocks() {
  mockUseDeleteWebEndpoint.mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useDeleteWebEndpoint>);

  mockUseCreateWebEndpoint.mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
    reset: vi.fn(),
  } as unknown as ReturnType<typeof useCreateWebEndpoint>);

  mockUseDevices.mockReturnValue({
    devices: [],
    totalCount: 0,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  });

  mockUseHasPermission.mockReturnValue(true);
  mockUseDebouncedValue.mockImplementation((v: unknown) => v);
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <WebEndpoints />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  setupDefaultMocks();
});

describe("WebEndpoints — pagination count / controls decoupling", () => {
  it("shows the endpoint count when totalCount > 0 and only one page exists", () => {
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [makeEndpoint("ep1.example.com")],
      totalCount: 1,
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(screen.getByText(/1 endpoint/i)).toBeInTheDocument();
  });

  it("hides Prev/Next controls when only one page exists", () => {
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [makeEndpoint("ep1.example.com")],
      totalCount: 1,
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /next page/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Prev/Next controls and the count when multiple pages exist", () => {
    const endpoints = Array.from({ length: 10 }, (_, i) =>
      makeEndpoint(`ep${i + 1}.example.com`),
    );
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: endpoints,
      totalCount: 15,
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(screen.getByText(/15 endpoints/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /previous page/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /next page/i }),
    ).toBeInTheDocument();
  });

  it("does not show the Pagination nav when there are no endpoints", () => {
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [],
      totalCount: 0,
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });

  it("does not flash a '0 endpoints' count while a search request is in-flight", () => {
    mockUseDebouncedValue.mockReturnValue("some-query");
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [],
      totalCount: 0,
      isLoading: true,
      error: null,
    });

    renderPage();

    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });
});

describe("WebEndpoints — URL hydration", () => {
  it("reads page=2 and search=myhost from the URL and passes them to useWebEndpoints", () => {
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [makeEndpoint("ep1.example.com")],
      totalCount: 1,
      isLoading: false,
      error: null,
    });

    renderPage(["/?page=2&search=myhost"]);

    expect(mockUseWebEndpoints).toHaveBeenCalledWith(
      expect.objectContaining({ page: 2, addressFilter: "myhost" }),
    );
  });

  it("falls back to page=1 and empty search when URL has no params", () => {
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [],
      totalCount: 0,
      isLoading: false,
      error: null,
    });

    renderPage(["/"]);

    expect(mockUseWebEndpoints).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, addressFilter: "" }),
    );
  });
});

describe("WebEndpoints — search resets page to 1", () => {
  it("resets to page=1 when a new search term is typed while on page 2", async () => {
    const user = userEvent.setup();

    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [makeEndpoint("ep1.example.com")],
      totalCount: 25,
      isLoading: false,
      error: null,
    });

    renderPage(["/?page=2"]);

    expect(mockUseWebEndpoints).toHaveBeenCalledWith(
      expect.objectContaining({ page: 2 }),
    );

    const searchInput = screen.getByPlaceholderText(/search by address/i);
    await user.type(searchInput, "x");

    expect(mockUseWebEndpoints).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 }),
    );
  });
});

describe("WebEndpoints — page change writes to URL", () => {
  it("calls useWebEndpoints with page=2 after clicking the Next page button", async () => {
    const user = userEvent.setup();

    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: Array.from({ length: 10 }, (_, i) =>
        makeEndpoint(`ep${i + 1}.example.com`),
      ),
      totalCount: 15,
      isLoading: false,
      error: null,
    });

    renderPage(["/"]);

    const nextButton = screen.getByRole("button", { name: /next page/i });
    await user.click(nextButton);

    expect(mockUseWebEndpoints).toHaveBeenCalledWith(
      expect.objectContaining({ page: 2 }),
    );
  });
});

async function openEndpointDrawer(user: ReturnType<typeof userEvent.setup>) {
  mockUseWebEndpoints.mockReturnValue({
    webEndpoints: [makeEndpoint("ep1.example.com")],
    totalCount: 1,
    isLoading: false,
    error: null,
  });

  renderPage();
  await user.click(screen.getByRole("button", { name: /new endpoint/i }));
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
