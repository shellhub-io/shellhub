import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { GetLicenseResponse } from "@/client/types.gen";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));

vi.mock("@/hooks/useAdminStats", () => ({
  useAdminStats: vi.fn(),
}));

import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAdminStats } from "@/hooks/useAdminStats";
import DeviceLimitBanner from "../DeviceLimitBanner";

// ── Typed mocks ───────────────────────────────────────────────────────────────

const mockUseAdminLicense = vi.mocked(useAdminLicense);
const mockUseAdminStats = vi.mocked(useAdminStats);

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
    dataUpdatedAt: Date.now(),
    ...overrides,
  } as unknown as UseQueryResult<LicenseData, Error>;
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
});
