import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import type { GetLicenseResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import LicenseBanner from "../LicenseBanner";

const mockGetLicense = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getLicense: mockGetLicense };
});

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

function renderBanner() {
  return render(<LicenseBanner />, { wrapper: createTestWrapper() });
}

async function waitForQuery() {
  await waitFor(() => {
    expect(mockGetLicense).toHaveBeenCalled();
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ isAdmin: true });
});

describe("LicenseBanner", () => {
  describe("visibility", () => {
    it("is hidden while the license check is in progress", () => {
      mockGetLicense.mockReturnValue(new Promise(() => {}));
      renderBanner();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("is hidden when the query is not enabled (non-admin)", async () => {
      useAuthStore.setState({ isAdmin: false } as never);
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense()));
      renderBanner();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
      });
      expect(mockGetLicense).not.toHaveBeenCalled();
    });

    it("is hidden when the query fails unexpectedly", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(500));
      renderBanner();
      await waitForQuery();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is hidden when license is valid", async () => {
      mockGetLicense.mockResolvedValue(mockSdkResponse(makeLicense()));
      renderBanner();
      await waitForQuery();
      await waitFor(() => {
        expect(screen.queryByRole("alert")).not.toBeInTheDocument();
        expect(screen.queryByRole("status")).not.toBeInTheDocument();
      });
    });

    it("is shown when no license is installed", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("is shown when license is expired", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: false })),
      );
      renderBanner();
      await waitFor(() => {
        expect(
          screen.getByText(/your license has expired\./i),
        ).toBeInTheDocument();
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("is shown when license is in the grace period", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: true })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByText(/grace period/i)).toBeInTheDocument();
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
    });

    it("is shown when license is about to expire", async () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(
          screen.getByText(/about to expire|expires in/i),
        ).toBeInTheDocument();
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
    });
  });

  describe("severity", () => {
    it("uses error (role=alert) when no license is installed", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("uses error (role=alert) when license is expired", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: false })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("uses warning (role=status) when license is in the grace period", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: true })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });

    it("uses warning (role=status) when license is about to expire", async () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("messages", () => {
    it("shows the no-license message", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
        expect(screen.getByText(/no license installed/i)).toBeInTheDocument();
      });
    });

    it("shows the expired message", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: false })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
        expect(
          screen.getByText(
            /your license has expired\. this instance won't function/i,
          ),
        ).toBeInTheDocument();
      });
    });

    it("shows the grace period message", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true, grace_period: true })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.getByText(/grace period/i)).toBeInTheDocument();
      });
    });

    it("shows days remaining when about to expire and days are known", async () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 1 * 86400;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.getByText(/expires in 1 day\b/i)).toBeInTheDocument();
      });
    });

    it("uses the plural form when more than one day remains", async () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.getByText(/expires in 5 days/i)).toBeInTheDocument();
      });
    });

    it("shows the fallback about-to-expire message when expires_at is not set", async () => {
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ about_to_expire: true, expires_at: -1 })),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
      });
    });

    it("shows fallback about-to-expire copy when expires_at is in the past", async () => {
      const expiredAt = Math.floor(Date.now() / 1000) - 1;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiredAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
      });
    });

    it("shows fallback about-to-expire copy when days would be zero", async () => {
      const nowSeconds = Math.floor(Date.now() / 1000);
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: nowSeconds }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
        expect(screen.queryByText(/expires in 0 day/i)).not.toBeInTheDocument();
        expect(screen.getByText(/is about to expire/i)).toBeInTheDocument();
      });
    });
  });

  describe("no CTA link", () => {
    it("never renders any link when no license is installed (error state)", async () => {
      mockGetLicense.mockRejectedValue(makeSdkError(400));
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/upload license/i)).not.toBeInTheDocument();
    });

    it("never renders any link when license is about to expire (warning state)", async () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 5 * 86400;
      mockGetLicense.mockResolvedValue(
        mockSdkResponse(
          makeLicense({ about_to_expire: true, expires_at: expiresAt }),
        ),
      );
      renderBanner();
      await waitFor(() => {
        expect(screen.getByRole("status")).toBeInTheDocument();
      });
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/upload license/i)).not.toBeInTheDocument();
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
