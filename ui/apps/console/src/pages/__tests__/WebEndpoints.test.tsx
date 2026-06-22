import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { Webendpoint } from "@/client/types.gen";
import WebEndpoints from "../WebEndpoints";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("@/hooks/useWebEndpoints");
vi.mock("@/hooks/useWebEndpointMutations");
vi.mock("@/hooks/useDevices");
vi.mock("@/hooks/useResetOnOpen");
vi.mock("@/hooks/useHasPermission");
vi.mock("@/hooks/useDebouncedValue");

import { useWebEndpoints } from "@/hooks/useWebEndpoints";
import { useDeleteWebEndpoint, useCreateWebEndpoint } from "@/hooks/useWebEndpointMutations";
import { useDevices } from "@/hooks/useDevices";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";

const mockUseWebEndpoints = vi.mocked(useWebEndpoints);
const mockUseDeleteWebEndpoint = vi.mocked(useDeleteWebEndpoint);
const mockUseCreateWebEndpoint = vi.mocked(useCreateWebEndpoint);
const mockUseDevices = vi.mocked(useDevices);
const mockUseHasPermission = vi.mocked(useHasPermission);
const mockUseDebouncedValue = vi.mocked(useDebouncedValue);

// Minimal fixture for the fields WebEndpoints renders; cast to the full
// generated type so the mocked hook return stays type-compatible.
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

function renderPage() {
  return render(
    <MemoryRouter>
      <WebEndpoints />
    </MemoryRouter>,
  );
}

/* ------------------------------------------------------------------ */
/* Tests                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  setupDefaultMocks();
});

describe("WebEndpoints — pagination count / controls decoupling", () => {
  it("shows the endpoint count when totalCount > 0 and only one page exists", () => {
    // One endpoint fits on a single page — totalPages = Math.ceil(1/10) = 1
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [makeEndpoint("ep1.example.com")],
      totalCount: 1,
      isLoading: false,
      error: null,
    });

    renderPage();

    // The count span MUST appear even though there is only 1 page
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

    // Prev/Next must NOT be present when totalPages === 1.
    // The buttons expose their accessible name via aria-label ("Previous/Next page").
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /next page/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Prev/Next controls and the count when multiple pages exist", () => {
    // 15 endpoints across 2 pages (perPage = 10)
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
    // The buttons expose their accessible name via aria-label ("Previous/Next page").
    expect(
      screen.getByRole("button", { name: /previous page/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /next page/i })).toBeInTheDocument();
  });

  it("does not show the Pagination nav when there are no endpoints (truly empty, no search)", () => {
    // isTrulyEmpty=true -> EmptyState renders; Pagination component is never mounted.
    mockUseWebEndpoints.mockReturnValue({
      webEndpoints: [],
      totalCount: 0,
      isLoading: false,
      error: null,
    });

    renderPage();

    // The EmptyState branch is taken — no pagination rendered at all.
    expect(screen.queryByText(/0 endpoints/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /previous page/i }),
    ).not.toBeInTheDocument();
  });

  it("does not flash a '0 endpoints' count while a search request is in-flight", () => {
    // isSearching=true, isLoading=true -> the Pagination component must be suppressed
    // so stale zero counts do not appear mid-request.
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
