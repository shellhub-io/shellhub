import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useSupportIdentifier } from "../useSupportIdentifier";

const mockGetNamespaceSupport = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getNamespaceSupport: mockGetNamespaceSupport };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useSupportIdentifier", () => {
  describe("when enabled=false", () => {
    it("never fires the query and returns null identifier", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", false),
      );

      expect(mockGetNamespaceSupport).not.toHaveBeenCalled();
      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("when tenantId is empty", () => {
    it("does not fire the query when tenantId is empty string", () => {
      renderHookWithClient(() => useSupportIdentifier("", true));

      expect(mockGetNamespaceSupport).not.toHaveBeenCalled();
    });

    it("does not fire the query when tenantId is null", () => {
      renderHookWithClient(() => useSupportIdentifier(null, true));

      expect(mockGetNamespaceSupport).not.toHaveBeenCalled();
    });

    it("returns null identifier when disabled by empty tenantId", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("", true),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe("when enabled with a valid tenant", () => {
    it("returns the identifier from the response", async () => {
      mockGetNamespaceSupport.mockResolvedValue(
        mockSdkResponse({ identifier: "abc123" }),
      );

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.identifier).toBe("abc123"));
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("retry policy", () => {
    it("retries the query exactly once on failure (transient blip recovery)", async () => {
      mockGetNamespaceSupport.mockRejectedValue(new Error("network error"));

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(mockGetNamespaceSupport).toHaveBeenCalledTimes(2);
    });
  });
});
