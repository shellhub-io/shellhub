import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { GetLicenseResponse } from "@/client/types.gen";
import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));

vi.mock("@/hooks/useAdminStats", () => ({
  useAdminStats: vi.fn(),
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
import { useAdminStats } from "@/hooks/useAdminStats";
import { getConfig } from "@/env";
import { getLicense } from "@/client";
import DeviceLimitBanner from "../DeviceLimitBanner";

// ── Typed mocks ───────────────────────────────────────────────────────────────

const mockUseAdminLicense = vi.mocked(useAdminLicense);
const mockUseAdminStats = vi.mocked(useAdminStats);
const mockGetConfig = vi.mocked(getConfig);
const mockGetLicense = vi.mocked(getLicense);

type LicenseData = GetLicenseResponse | null;

function makeLicenseQueryResult(
  overrides: {
    data?: LicenseData;
    isLoading?: boolean;
    isError?: boolean;
    dataUpdatedAt?: number;
  } = {},
) {
  return {
    data: undefined,
    isLoading: false,
    isError: false,
    isExpired: false,
    installedLicense: null,
    dataUpdatedAt: Date.now(),
    ...overrides,
  } as unknown as ReturnType<typeof useAdminLicense>;
}

function makeStatsResult(
  overrides: {
    stats?: { registered_devices?: number } | undefined;
    isLoading?: boolean;
    isError?: boolean;
    error?: Error | null;
  } = {},
) {
  return {
    stats: undefined,
    isLoading: false,
    isError: false,
    error: null,
    ...overrides,
  };
}

function makeLicense(
  devicesLimit: number,
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
      devices: devicesLimit,
      session_recording: true,
      firewall_rules: true,
      billing: false,
      login_link: false,
      reports: false,
    },
    ...overrides,
  } as GetLicenseResponse;
}

// ── Setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
  // Default: enterprise non-cloud config
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  // Default: admin with a license allowing 100 devices
  mockUseAdminLicense.mockReturnValue(
    makeLicenseQueryResult({ data: makeLicense(100) }),
  );
  mockUseAdminStats.mockReturnValue(
    makeStatsResult({ stats: { registered_devices: 50 } }),
  );
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("DeviceLimitBanner", () => {
  describe("severity: over limit", () => {
    it("shows alert (role=alert) with over-limit copy when registered >= cap", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 100 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText(/you've reached your licensed device limit/i),
      ).toBeInTheDocument();
      expect(
        screen.getByText(/contact the ShellHub team/i),
      ).toBeInTheDocument();
    });

    it("shows RED (role=alert) when cap=10 and registered=10", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(10) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 10 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("shows RED (role=alert) when cap=0 and registered=0 (cap===0 -> over)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(0) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 0 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });
  });

  describe("severity: approaching limit", () => {
    it("shows status (role=status) with approaching copy when at 90% but under cap", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 90 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(
        screen.getByText(/you're approaching your licensed device limit/i),
      ).toBeInTheDocument();
      expect(
        screen.getByText(/contact the ShellHub team/i),
      ).toBeInTheDocument();
    });

    it("shows YELLOW (role=status) when cap=10 and registered=9 (90% boundary)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(10) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 9 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("visibility guards", () => {
    it("is absent when cap=10 and registered=8 (80% — below threshold)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(10) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 8 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when features.devices === -1 (unlimited)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(-1) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 9999 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when registered_devices is undefined", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: undefined } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when non-admin (hooks return undefined data)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: undefined }),
      );
      mockUseAdminStats.mockReturnValue(makeStatsResult({ stats: undefined }));
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent while license is loading (even when over-limit data is present)", () => {
      // data present + over-limit would normally show the banner; isLoading must suppress it.
      // Using data: undefined would hide the banner via the license != null guard, not the
      // licenseLoading guard — so the guard under test would never actually be exercised.
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100), isLoading: true }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 100 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when license is null (no license installed)", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: null }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when useAdminStats isError is true", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100) }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 100 }, isError: true }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent while stats are loading", () => {
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({
          stats: { registered_devices: 100 },
          isLoading: true,
        }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when the license query errored", () => {
      mockUseAdminLicense.mockReturnValue(
        makeLicenseQueryResult({ data: makeLicense(100), isError: true }),
      );
      mockUseAdminStats.mockReturnValue(
        makeStatsResult({ stats: { registered_devices: 100 } }),
      );
      render(<DeviceLimitBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });

  describe("cloud deployment", () => {
    it("is hidden when cloud=true and admin=true (banner suppressed, getLicense never fires)", async () => {
      // Do NOT mock useAdminLicense for this test: restore the real implementation
      // to verify the cloud bypass path end-to-end through DeviceLimitBanner.
      // When cloud=true the query is disabled, data stays undefined, and the banner
      // must stay hidden regardless of getLicense responses.
      const { useAdminLicense: real } = await vi.importActual<
        typeof import("@/hooks/useAdminLicense")
      >("@/hooks/useAdminLicense");
      mockUseAdminLicense.mockImplementation(real);

      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      useAuthStore.setState({ isAdmin: true } as never);
      // getLicense must never fire on cloud deployments.
      mockGetLicense.mockRejectedValue({ status: 400 });

      // useAdminStats is still mocked from beforeEach; stats are irrelevant here
      // because the license guard (license != null) will suppress the banner first.

      render(
        <QueryClientProvider
          client={
            new QueryClient({
              defaultOptions: { queries: { retry: false, retryDelay: 0 } },
            })
          }
        >
          <DeviceLimitBanner />
        </QueryClientProvider>,
      );

      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
      expect(mockGetLicense).not.toHaveBeenCalled();
    });
  });
});
