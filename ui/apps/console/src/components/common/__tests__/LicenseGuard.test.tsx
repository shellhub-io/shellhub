import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import type { ComponentType, ReactNode } from "react";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import type { GetLicenseResponse } from "@/client";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import LicenseGuard from "../LicenseGuard";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getLicense: vi.fn(),
  }),
);

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

function renderGuard(Wrapper?: ComponentType<{ children: ReactNode }>) {
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

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ isAdmin: true });
});

describe("LicenseGuard", () => {
  describe("isLoading — shows PageLoader", () => {
    it("renders a loading indicator while the license check is in progress", () => {
      sdk.getLicense.mockReturnValue(new Promise(() => {}));
      renderGuard(createTestWrapper());
      expect(screen.getByText("Checking license...")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });

    it("does not render the Outlet while loading", () => {
      sdk.getLicense.mockReturnValue(new Promise(() => {}));
      renderGuard(createTestWrapper());
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isError — redirects to /admin/license", () => {
    it("navigates to the license page when the query errors", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(500));
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("no license — redirects to /admin/license", () => {
    it("navigates to the license page when no license is installed", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isExpired — redirects to /admin/license", () => {
    it("navigates to the license page when the license is expired", async () => {
      sdk.getLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true })),
      );
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("valid license — renders Outlet", () => {
    it("renders the child route when the license is valid", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(makeLicense()));
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("protected content")).toBeInTheDocument();
      });
      expect(screen.queryByText("license page")).not.toBeInTheDocument();
    });
  });

  describe("cloud deployment", () => {
    it("renders the Outlet without calling getLicense when cloud=true", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      sdk.getLicense.mockRejectedValue(makeSdkError(400));

      renderGuard(createTestWrapper());

      await waitFor(() => {
        expect(screen.getByText("protected content")).toBeInTheDocument();
      });
      expect(sdk.getLicense).not.toHaveBeenCalled();
    });
  });
});
