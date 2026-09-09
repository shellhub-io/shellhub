import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import type { GetLicense200 as GetLicenseResponse } from "@/client/model";
import { useAuthStore } from "@/stores/authStore";
import LicenseBanner from "../LicenseBanner";

const DAY = 86400;

const mockGetConfig = vi.mocked(getConfig);

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

function setLicense(overrides: Partial<GetLicenseResponse> = {}) {
  server.use(
    http.get("*/admin/api/license", () =>
      HttpResponse.json(makeLicense(overrides)),
    ),
  );
}

function setNoLicense() {
  server.use(
    http.get("*/admin/api/license", () =>
      HttpResponse.json({}, { status: 400 }),
    ),
  );
}

function nowSeconds() {
  return Math.floor(Date.now() / 1000);
}

function renderBanner() {
  return render(<LicenseBanner />, { wrapper: createTestWrapper() });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ isAdmin: true });
});

describe("LicenseBanner", () => {
  describe("hidden states", () => {
    it("is hidden while the license check is in progress", () => {
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it.each([
      [
        "the query is not enabled (non-admin)",
        () => {
          useAuthStore.setState({ isAdmin: false });
          setLicense();
        },
      ],
      [
        "the query fails unexpectedly",
        () =>
          server.use(
            http.get("*/admin/api/license", () =>
              HttpResponse.json({}, { status: 500 }),
            ),
          ),
      ],
      ["the license is valid", () => setLicense()],
      [
        "the edition is cloud, so getLicense never fires",
        () =>
          mockGetConfig.mockReturnValue({
            ...defaultConfig,
            edition: "cloud",
          }),
      ],
    ])("is hidden when %s", async (_label, arrange) => {
      arrange();
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });
  });

  describe("severity and message", () => {
    it.each([
      ["no license is installed", setNoLicense, "alert", /no license installed/i],
      [
        "the license is expired",
        () => setLicense({ expired: true, grace_period: false }),
        "alert",
        /your license has expired\. this instance won't function/i,
      ],
      [
        "the license is in the grace period",
        () => setLicense({ expired: true, grace_period: true }),
        "status",
        /grace period/i,
      ],
      [
        "one day remains",
        () =>
          setLicense({ about_to_expire: true, expires_at: nowSeconds() + DAY }),
        "status",
        /expires in 1 day\b/i,
      ],
      [
        "several days remain",
        () =>
          setLicense({
            about_to_expire: true,
            expires_at: nowSeconds() + 5 * DAY,
          }),
        "status",
        /expires in 5 days/i,
      ],
      [
        "expires_at is not set",
        () => setLicense({ about_to_expire: true, expires_at: -1 }),
        "status",
        /is about to expire/i,
      ],
      [
        "expires_at is already in the past",
        () => setLicense({ about_to_expire: true, expires_at: nowSeconds() - 1 }),
        "status",
        /is about to expire/i,
      ],
      [
        "the remaining days would round to zero",
        () => setLicense({ about_to_expire: true, expires_at: nowSeconds() }),
        "status",
        /is about to expire/i,
      ],
    ] as const)("when %s it is a %s reading '%s'", async (
      _label,
      arrange,
      role,
      message,
    ) => {
      arrange();
      renderBanner();

      await waitFor(() => expect(screen.getByRole(role)).toBeInTheDocument());
      expect(screen.getByText(message)).toBeInTheDocument();
      expect(
        screen.queryByRole(role === "alert" ? "status" : "alert"),
      ).not.toBeInTheDocument();
      expect(screen.queryByText(/expires in 0 day/i)).not.toBeInTheDocument();
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/upload license/i)).not.toBeInTheDocument();
    });
  });
});
