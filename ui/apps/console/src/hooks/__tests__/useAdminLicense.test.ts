import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn() };
});

vi.mock("@/client", () => ({
  getLicense: vi.fn(),
}));

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getLicenseQueryKey: vi.fn(() => ["getLicense"]),
}));

vi.mock("@/api/errors", () => ({
  isSdkError: vi.fn(
    (err: unknown): err is { status: number } =>
      typeof err === "object" && err !== null && "status" in err,
  ),
}));

import { getConfig } from "@/env";
import { getLicense } from "@/client";
import { useAdminLicense } from "../useAdminLicense";

const mockGetConfig = vi.mocked(getConfig);
const mockGetLicense = vi.mocked(getLicense);

// ── Helpers ───────────────────────────────────────────────────────────────────

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
    },
  });

  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

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
  // Default: enterprise admin, non-cloud
  useAuthStore.setState({ isAdmin: true } as never);
  mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: false });
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("useAdminLicense", () => {
  describe("enterprise admin — valid license", () => {
    it("calls getLicense and returns installedLicense", async () => {
      const license = makeLicense();
      mockGetLicense.mockResolvedValue({ data: license } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(mockGetLicense).toHaveBeenCalledTimes(1);
      expect(result.current.installedLicense).toEqual(license);
    });

    it("sets isExpired to false when license is not expired", async () => {
      mockGetLicense.mockResolvedValue({
        data: makeLicense({ expired: false }),
      } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("enterprise admin — 400 (no license stored)", () => {
    it("normalizes 400 to installedLicense null", async () => {
      mockGetLicense.mockRejectedValue({ status: 400 });

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.installedLicense).toBeNull();
    });

    it("sets isExpired to true when no license is installed", async () => {
      mockGetLicense.mockRejectedValue({ status: 400 });

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });
  });

  describe("enterprise admin — expired license", () => {
    it("sets isExpired to true when license.expired is true", async () => {
      mockGetLicense.mockResolvedValue({
        data: makeLicense({ expired: true }),
      } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });

    it("keeps isExpired false for a non-expired license still in its grace period", async () => {
      mockGetLicense.mockResolvedValue({
        data: makeLicense({ expired: false, grace_period: true }),
      } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("cloud admin — bypass", () => {
    it("does NOT call getLicense on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: true });

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      // Give React Query a chance to fire (it should not)
      await new Promise((r) => setTimeout(r, 20));

      expect(mockGetLicense).not.toHaveBeenCalled();
      expect(result.current.isLoading).toBe(false);
    });

    it("returns isExpired false on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, cloud: true });

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await new Promise((r) => setTimeout(r, 20));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("non-admin on enterprise", () => {
    it("does NOT call getLicense when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await new Promise((r) => setTimeout(r, 20));

      expect(mockGetLicense).not.toHaveBeenCalled();
      expect(result.current.isLoading).toBe(false);
    });

    it("returns isExpired false when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminLicense(), {
        wrapper: createWrapper(),
      });

      await new Promise((r) => setTimeout(r, 20));
      expect(result.current.isExpired).toBe(false);
    });
  });
});
