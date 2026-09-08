import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import type { GetLicense200 as GetLicenseResponse } from "@/client/model";
import { useAuthStore } from "@/stores/authStore";
import DeviceLimitBanner from "../DeviceLimitBanner";

const mockGetConfig = vi.mocked(getConfig);

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

function setHandlers(
  license: GetLicenseResponse,
  stats: Record<string, unknown>,
) {
  server.use(
    http.get("*/admin/api/license", () => HttpResponse.json(license)),
    http.get("*/api/stats", () => HttpResponse.json(stats)),
  );
}

function renderBanner() {
  return render(<DeviceLimitBanner />, { wrapper: createTestWrapper() });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ isAdmin: true });
  setHandlers(makeLicense(100), { registered_devices: 50 });
});

describe("DeviceLimitBanner", () => {
  describe("severity: over limit", () => {
    it("shows alert (role=alert) with over-limit copy when registered >= cap", async () => {
      setHandlers(makeLicense(100), { registered_devices: 100 });
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
        expect(
          screen.getByText(/you've reached your licensed device limit/i),
        ).toBeInTheDocument();
        expect(
          screen.getByText(/contact the ShellHub team/i),
        ).toBeInTheDocument();
      });
    });

    it("shows RED (role=alert) when cap=10 and registered=10", async () => {
      setHandlers(makeLicense(10), { registered_devices: 10 });
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("shows RED (role=alert) when cap=0 and registered=0 (cap===0 -> over)", async () => {
      setHandlers(makeLicense(0), { registered_devices: 0 });
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });
  });

  describe("severity: approaching limit", () => {
    it("shows status (role=status) with approaching copy when at 90% but under cap", async () => {
      setHandlers(makeLicense(100), { registered_devices: 90 });
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(
          screen.getByText(/you're approaching your licensed device limit/i),
        ).toBeInTheDocument();
        expect(
          screen.getByText(/contact the ShellHub team/i),
        ).toBeInTheDocument();
      });
    });

    it("shows YELLOW (role=status) when cap=10 and registered=9 (90% boundary)", async () => {
      setHandlers(makeLicense(10), { registered_devices: 9 });
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("visibility guards", () => {
    it("is absent when cap=10 and registered=8 (80% — below threshold)", async () => {
      setHandlers(makeLicense(10), { registered_devices: 8 });
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when features.devices === -1 (unlimited)", async () => {
      setHandlers(makeLicense(-1), { registered_devices: 9999 });
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when registered_devices is undefined", async () => {
      setHandlers(makeLicense(100), {});
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when non-admin (queries disabled)", async () => {
      useAuthStore.setState({ isAdmin: false });
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent while license is loading", () => {
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when no license is installed", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when useAdminStats errors", async () => {
      server.use(
        http.get("*/api/stats", () => HttpResponse.json({}, { status: 500 })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent while stats are loading", () => {
      server.use(http.get("*/api/stats", () => new Promise(() => {})));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when the license query errored", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });
  });

  describe("cloud deployment", () => {
    it("is hidden when cloud=true and admin=true (getLicense never fires)", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });

      renderBanner();

      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });
  });
});
