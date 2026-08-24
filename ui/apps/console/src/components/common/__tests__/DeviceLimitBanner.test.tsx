import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import type { GetLicenseResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import DeviceLimitBanner from "../DeviceLimitBanner";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getLicense: vi.fn(),
    getStats: vi.fn(),
  }),
);

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

function renderBanner() {
  return render(<DeviceLimitBanner />, { wrapper: createTestWrapper() });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ isAdmin: true });
  sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
  sdk.getStats.mockResolvedValue(mockSdkResponse({ registered_devices: 50 }));
});

async function waitForQueries() {
  await waitFor(() => {
    expect(sdk.getLicense).toHaveBeenCalled();
    expect(sdk.getStats).toHaveBeenCalled();
  });
}

describe("DeviceLimitBanner", () => {
  describe("severity: over limit", () => {
    it("shows alert (role=alert) with over-limit copy when registered >= cap", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 100 }),
      );
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
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 10 }),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("shows RED (role=alert) when cap=0 and registered=0 (cap===0 -> over)", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(0)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 0 }),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });
  });

  describe("severity: approaching limit", () => {
    it("shows status (role=status) with approaching copy when at 90% but under cap", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 90 }),
      );
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
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 9 }),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("visibility guards", () => {
    it("is absent when cap=10 and registered=8 (80% — below threshold)", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 8 }),
      );
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when features.devices === -1 (unlimited)", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(-1)));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 9999 }),
      );
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when registered_devices is undefined", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      sdk.getStats.mockResolvedValue(mockSdkResponse({}));
      renderBanner();
      await waitForQueries();
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
      expect(sdk.getLicense).not.toHaveBeenCalled();
      expect(sdk.getStats).not.toHaveBeenCalled();
    });

    it("is absent while license is loading", () => {
      sdk.getLicense.mockReturnValue(new Promise(() => {}));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 100 }),
      );
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when no license is installed", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when useAdminStats errors", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      sdk.getStats.mockRejectedValue(makeSdkError(500));
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent while stats are loading", () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      sdk.getStats.mockReturnValue(new Promise(() => {}));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when the license query errored", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(500));
      sdk.getStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 100 }),
      );
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });
  });

  describe("cloud deployment", () => {
    it("is hidden when cloud=true and admin=true (getLicense never fires)", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      sdk.getLicense.mockRejectedValue(makeSdkError(400));

      renderBanner();

      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
      expect(sdk.getLicense).not.toHaveBeenCalled();
    });
  });
});
