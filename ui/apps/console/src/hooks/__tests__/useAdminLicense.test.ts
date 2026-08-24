import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getLicense: vi.fn(),
  }),
);

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
      sdk.getLicense.mockResolvedValue(mockSdkResponse(license));

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(sdk.getLicense).toHaveBeenCalledTimes(1);
      expect(result.current.installedLicense).toEqual(license);
    });

    it("sets isExpired to false when license is not expired", async () => {
      sdk.getLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: false })),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("enterprise admin — 400 (no license stored)", () => {
    it("normalizes 400 to installedLicense null", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.installedLicense).toBeNull();
    });

    it("sets isExpired to true when no license is installed", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });
  });

  describe("enterprise admin — expired license", () => {
    it("sets isExpired to true when license.expired is true", async () => {
      sdk.getLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: true })),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(true);
    });

    it("keeps isExpired false for a non-expired license still in its grace period", async () => {
      sdk.getLicense.mockResolvedValue(
        mockSdkResponse(makeLicense({ expired: false, grace_period: true })),
      );

      const { result } = renderHookWithClient(() => useAdminLicense());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("cloud admin — bypass", () => {
    it("does NOT call getLicense on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await new Promise((r) => setTimeout(r, 20));

      expect(sdk.getLicense).not.toHaveBeenCalled();
      expect(result.current.isLoading).toBe(false);
    });

    it("returns isExpired false on cloud deployments", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await new Promise((r) => setTimeout(r, 20));
      expect(result.current.isExpired).toBe(false);
    });
  });

  describe("non-admin on enterprise", () => {
    it("does NOT call getLicense when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await new Promise((r) => setTimeout(r, 20));

      expect(sdk.getLicense).not.toHaveBeenCalled();
      expect(result.current.isLoading).toBe(false);
    });

    it("returns isExpired false when user is not admin", async () => {
      useAuthStore.setState({ isAdmin: false });

      const { result } = renderHookWithClient(() => useAdminLicense());

      await new Promise((r) => setTimeout(r, 20));
      expect(result.current.isExpired).toBe(false);
    });
  });
});
