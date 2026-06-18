import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import type { UseQueryResult } from "@tanstack/react-query";
import type { GetLicenseResponse } from "@/client/types.gen";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));

import { useAdminLicense } from "@/hooks/useAdminLicense";
import LicenseBanner from "../LicenseBanner";

// ── Typed mocks ───────────────────────────────────────────────────────────────

const mockUseAdminLicense = vi.mocked(useAdminLicense);

type LicenseData = GetLicenseResponse | null;

function makeQueryResult(
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

// ── Setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
  // Default: admin with no license installed
  mockUseAdminLicense.mockReturnValue(makeQueryResult({ data: null }));
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("LicenseBanner", () => {
  describe("visibility", () => {
    it("is hidden while the license check is in progress", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({ data: null, isLoading: true }),
      );
      render(<LicenseBanner />);
      // Both roles must be absent — checking only one would miss regressions that
      // make the banner visible but switch it from error to warning severity.
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is hidden when the query is not enabled (non-admin, data is undefined)", () => {
      mockUseAdminLicense.mockReturnValue(makeQueryResult({ data: undefined }));
      render(<LicenseBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is hidden when the query fails unexpectedly", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({ data: null, isError: true }),
      );
      render(<LicenseBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is hidden when license is valid", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({ data: makeLicense() }),
      );
      render(<LicenseBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is shown when no license is installed", () => {
      render(<LicenseBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("is shown when license is expired", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: false }),
        }),
      );
      render(<LicenseBanner />);
      expect(
        screen.getByText(/your license has expired\./i),
      ).toBeInTheDocument();
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("is shown when license is in the grace period", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: true }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByText(/grace period/i)).toBeInTheDocument();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("is shown when license is about to expire", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true }),
        }),
      );
      render(<LicenseBanner />);
      expect(
        screen.getByText(/about to expire|expires in/i),
      ).toBeInTheDocument();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("severity", () => {
    it("uses error (role=alert) when no license is installed", () => {
      render(<LicenseBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("uses error (role=alert) when license is expired", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: false }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("uses warning (role=status) when license is in the grace period", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: true }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("uses warning (role=status) when license is about to expire", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("messages", () => {
    it("shows the no-license message", () => {
      render(<LicenseBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/no license installed/i)).toBeInTheDocument();
    });

    it("shows the expired message", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: false }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText(
          /your license has expired\. this instance won't function/i,
        ),
      ).toBeInTheDocument();
    });

    it("shows the grace period message", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ expired: true, grace_period: true }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/grace period/i)).toBeInTheDocument();
    });

    it("shows days remaining when about to expire and days are known", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 1 * 86400;
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/expires in 1 day\b/i)).toBeInTheDocument();
    });

    it("uses the plural form when more than one day remains", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/expires in 5 days/i)).toBeInTheDocument();
    });

    it("shows the fallback about-to-expire message when expires_at is not set", () => {
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: -1 }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
    });

    it("shows fallback about-to-expire copy when expires_at is in the past and about_to_expire is true (negative days)", () => {
      // expires_at is in the past: daysUntilExpiration will be < 0
      const expiredAt = Math.floor(Date.now() / 1000) - 1;
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: expiredAt }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
    });

    it("shows fallback about-to-expire copy when Math.ceil yields exactly 0 (days===0 boundary)", () => {
      // expires_at exactly equal to dataUpdatedAt/1000 → Math.ceil(0/86400) = 0
      // The guard `days > 0 ? days : null` must return null so the fallback is shown.
      const now = Date.now();
      const nowSeconds = Math.floor(now / 1000);
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: nowSeconds }),
          dataUpdatedAt: now,
        }),
      );
      render(<LicenseBanner />);
      // days = Math.ceil((nowSeconds - now/1000) / 86400) = Math.ceil(0) = 0
      // → null → fallback copy shown, NOT "expires in 0 days"
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.queryByText(/expires in 0 day/i)).not.toBeInTheDocument();
      expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
    });
  });

  describe("no CTA link", () => {
    it("never renders any link when no license is installed (error state)", () => {
      render(<LicenseBanner />);
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/upload license/i)).not.toBeInTheDocument();
    });

    it("never renders any link when license is about to expire (warning state)", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockUseAdminLicense.mockReturnValue(
        makeQueryResult({
          data: makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        }),
      );
      render(<LicenseBanner />);
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/upload license/i)).not.toBeInTheDocument();
    });
  });
});
