import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";

import { useAdminLicense } from "../useAdminLicense";

const mockGetConfig = vi.mocked(getConfig);

function makeLicense(overrides: Record<string, unknown> = {}) {
  return {
    id: "license-1",
    expired: false,
    about_to_expire: false,
    grace_period: false,
    issued_at: 0,
    starts_at: 0,
    expires_at: -1,
    allowed_regions: [],
    customer: {},
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  mockGetConfig.mockReturnValue({ ...defaultConfig });
});

describe("useAdminLicense", () => {
  describe("enterprise admin — valid license", () => {
    it("calls getLicense and returns installedLicense", async () => {
      const license = makeLicense();
      server.use(
        http.get("*/admin/api/license", () => HttpResponse.json(license)),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(result.current.installedLicense).toEqual(license);
    });

    it("sets isExpired to false when license is not expired", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json(makeLicense({ expired: false })),
        ),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("enterprise admin — 400 (no license stored)", () => {
    it("normalizes 400 to installedLicense null", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.installedLicense).toBeNull();
    });

    it("sets isExpired to true when no license is installed", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });
  });

  describe("enterprise admin — expired license", () => {
    it("sets isExpired to true when license.expired is true", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json(makeLicense({ expired: true })),
        ),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });

    it("keeps isExpired false for a non-expired license still in its grace period", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json(
            makeLicense({ expired: false, grace_period: true }),
          ),
        ),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("cloud admin — bypass", () => {
    it("does NOT call getLicense on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      let called = false;
      server.use(
        http.get("*/admin/api/license", () => {
          called = true;
          return HttpResponse.json(makeLicense());
        }),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => {
        expect(called).toBe(false);
        expect(result.current.isLoading).toBe(false);
      });
    });

    it("returns isExpired false on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() =>
        expect(result.current.isExpired).toBe(false),
      );
    });
  });

  describe("non-admin on enterprise", () => {
    it("does NOT call getLicense when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false });
      let called = false;
      server.use(
        http.get("*/admin/api/license", () => {
          called = true;
          return HttpResponse.json(makeLicense());
        }),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => {
        expect(called).toBe(false);
        expect(result.current.isLoading).toBe(false);
      });
    });

    it("returns isExpired false when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() =>
        expect(result.current.isExpired).toBe(false),
      );
    });
  });
});
