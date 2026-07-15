import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import type { GetLicenseResponse } from "@/client/types.gen";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));
vi.mock("@/client", () => ({
  getLicense: vi.fn(),
}));

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getLicenseQueryKey: vi.fn(() => ["getLicense"]),
}));

vi.mock("@/api/errors", () => ({
  isSdkError: vi.fn(
    (err: unknown): err is { status: number } =>
      typeof err === "object" && err !== null && "status" in err,
  ),
}));

import { useAdminLicense } from "@/hooks/useAdminLicense";
import { getConfig } from "@/env";
import { getLicense } from "@/client";
import LicenseGuard from "../LicenseGuard";

// ── Typed mocks ───────────────────────────────────────────────────────────────

const mockUseAdminLicense = vi.mocked(useAdminLicense);
const mockGetConfig = vi.mocked(getConfig);
const mockGetLicense = vi.mocked(getLicense);

// ── Helpers ───────────────────────────────────────────────────────────────────

type LicenseData = GetLicenseResponse | null;

function makeLicense(
  overrides: Partial<GetLicenseResponse> = {},
): GetLicenseResponse {
  return {
    expired: false,
    grace_period: false,
    about_to_expire: false,
    expires_at: 9999999999,
    issued_at: 0,
    starts_at: 0,
    allowed_regions: [],
    customer: { id: "c1", name: "Acme", email: "a@b.com", company: "Acme" },
    features: {
      devices: -1,
      session_recording: true,
      firewall_rules: true,
      billing: false,
      login_link: false,
      reports: false,
    },
    ...overrides,
  } as GetLicenseResponse;
}

type HookReturn = ReturnType<typeof useAdminLicense>;

function makeHookReturn(
  overrides: {
    isLoading?: boolean;
    isError?: boolean;
    isExpired?: boolean;
    data?: LicenseData;
  } = {},
): HookReturn {
  return {
    data: undefined,
    isLoading: false,
    isError: false,
    isExpired: false,
    installedLicense: null,
    ...overrides,
  } as unknown as HookReturn;
}

function createQueryWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false, retryDelay: 0 } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

function renderGuard(
  Wrapper?: React.ComponentType<{ children: React.ReactNode }>,
) {
  const ui = (
    <MemoryRouter initialEntries={["/admin/dashboard"]}>
      <Routes>
        <Route element={<LicenseGuard />}>
          <Route
            path="/admin/dashboard"
            element={<div>protected content</div>}
          />
        </Route>
        <Route path="/admin/license" element={<div>license page</div>} />
      </Routes>
    </MemoryRouter>
  );
  return render(Wrapper ? <Wrapper>{ui}</Wrapper> : ui);
}

// ── Setup / teardown ──────────────────────────────────────────────────────────

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  mockUseAdminLicense.mockReturnValue(makeHookReturn());
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("LicenseGuard", () => {
  describe("isLoading — shows PageLoader", () => {
    it("renders a loading indicator while the license check is in progress", () => {
      mockUseAdminLicense.mockReturnValue(makeHookReturn({ isLoading: true }));
      renderGuard();
      expect(screen.getByText("Checking license...")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });

    it("does not render the Outlet while loading", () => {
      mockUseAdminLicense.mockReturnValue(makeHookReturn({ isLoading: true }));
      renderGuard();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isError — redirects to /admin/license", () => {
    it("navigates to the license page when the query errors", () => {
      mockUseAdminLicense.mockReturnValue(makeHookReturn({ isError: true }));
      renderGuard();
      expect(screen.getByText("license page")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isExpired — redirects to /admin/license", () => {
    it("navigates to the license page when the hook reports the license is expired", () => {
      // data contains a technically valid (unexpired) license, but the hook has
      // already derived isExpired=true (e.g. grace_period scenario).
      // The current component re-derives expiry locally from data and renders
      // the Outlet instead of redirecting — this test documents the desired behavior.
      mockUseAdminLicense.mockReturnValue(
        makeHookReturn({
          isExpired: true,
          data: makeLicense({ expired: false, grace_period: false }),
        }),
      );
      renderGuard();
      expect(screen.getByText("license page")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("valid license — renders Outlet", () => {
    it("renders the child route when all flags are false", () => {
      mockUseAdminLicense.mockReturnValue(
        makeHookReturn({ data: makeLicense() }),
      );
      renderGuard();
      expect(screen.getByText("protected content")).toBeInTheDocument();
      expect(screen.queryByText("license page")).not.toBeInTheDocument();
    });
  });

  describe("cloud real-hook path", () => {
    it("renders the Outlet without calling getLicense when cloud=true and admin=true", async () => {
      // Do NOT mock useAdminLicense for this test: restore the real implementation
      // to verify the cloud bypass path end-to-end through LicenseGuard.
      // When cloud=true the query is disabled, so getLicense must never fire,
      // isExpired stays false, and the guard must pass through to the Outlet.
      const { useAdminLicense: real } = await vi.importActual<
        typeof import("@/hooks/useAdminLicense")
      >("@/hooks/useAdminLicense");
      mockUseAdminLicense.mockImplementation(real);

      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      useAuthStore.setState({ isAdmin: true } as never);
      // getLicense must never fire on cloud deployments.
      mockGetLicense.mockRejectedValue({ status: 400 });

      renderGuard(createQueryWrapper());

      await waitFor(() => {
        expect(screen.getByText("protected content")).toBeInTheDocument();
      });
      expect(mockGetLicense).not.toHaveBeenCalled();
    });
  });
});
