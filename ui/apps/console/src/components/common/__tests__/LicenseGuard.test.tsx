import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import type { ComponentType, ReactNode } from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import type { GetLicense200 as GetLicenseResponse } from "@/client/model";
import LicenseGuard from "../LicenseGuard";

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
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderGuard(createTestWrapper());
      expect(screen.getByText("Checking license...")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });

    it("does not render the Outlet while loading", () => {
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderGuard(createTestWrapper());
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isError — redirects to /admin/license", () => {
    it("navigates to the license page when the query errors", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("no license — redirects to /admin/license", () => {
    it("navigates to the license page when no license is installed", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isExpired — redirects to /admin/license", () => {
    it("navigates to the license page when the license is expired", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json(makeLicense({ expired: true })),
        ),
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
      server.use(
        http.get("*/admin/api/license", () => HttpResponse.json(makeLicense())),
      );
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
      let called = false;
      server.use(
        http.get("*/admin/api/license", () => {
          called = true;
          return HttpResponse.json({}, { status: 400 });
        }),
      );

      renderGuard(createTestWrapper());

      await waitFor(() => {
        expect(screen.getByText("protected content")).toBeInTheDocument();
      });
      expect(called).toBe(false);
    });
  });
});
