import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import type { GetLicenseResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import DeviceLimitBanner from "../DeviceLimitBanner";

const mockGetLicense = vi.hoisted(() => vi.fn());
const mockGetStats = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getLicense: mockGetLicense, getStats: mockGetStats };
});

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
  mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
  mockGetStats.mockResolvedValue(mockSdkResponse({ registered_devices: 50 }));
});

async function waitForQueries() {
  await waitFor(() => {
    expect(mockGetLicense).toHaveBeenCalled();
    expect(mockGetStats).toHaveBeenCalled();
  });
}

describe("DeviceLimitBanner", () => {
  describe("severity: over limit", () => {
    it("shows alert (role=alert) with over-limit copy when registered >= cap", async () => {
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      mockGetStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 10 }),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("shows RED (role=alert) when cap=0 and registered=0 (cap===0 -> over)", async () => {
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(0)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(10)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(-1)));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      mockGetStats.mockResolvedValue(mockSdkResponse({}));
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when non-admin (queries disabled)", async () => {
      useAuthStore.setState({ isAdmin: false } as never);
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
      expect(mockGetLicense).not.toHaveBeenCalled();
      expect(mockGetStats).not.toHaveBeenCalled();
    });

    it("is absent while license is loading", () => {
      mockGetLicense.mockReturnValue(new Promise(() => {}));
      mockGetStats.mockResolvedValue(
        mockSdkResponse({ registered_devices: 100 }),
      );
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when no license is installed", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent when useAdminStats errors", async () => {
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      mockGetStats.mockRejectedValue(makeSdkError(500));
      renderBanner();
      await waitForQueries();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is absent while stats are loading", () => {
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense(100)));
      mockGetStats.mockReturnValue(new Promise(() => {}));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is absent when the license query errored", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(500));
      mockGetStats.mockResolvedValue(
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
      mockGetLicense.mockRejectedValue(makeSdkError(400));

      renderBanner();

      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
      expect(mockGetLicense).not.toHaveBeenCalled();
    });
  });
});
