import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";

const mockGetNamespaceSupportFn = vi.fn();

vi.mock("@/client", () => ({
  getNamespaceSupportOptions: vi.fn(() => ({
    queryKey: [{ _id: "getNamespaceSupport" }],
    queryFn: mockGetNamespaceSupportFn,
  })),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

async function importHook() {
  return await import("../useSupportIdentifier");
}

describe("useSupportIdentifier", () => {
  describe("when enabled=false", () => {
    it("never fires the query and returns null identifier", async () => {
      const { useSupportIdentifier } = await importHook();

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", false),
      );

      expect(mockGetNamespaceSupportFn).not.toHaveBeenCalled();
      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("when tenantId is empty", () => {
    it("does not fire the query when tenantId is empty string", async () => {
      const { useSupportIdentifier } = await importHook();

      renderHookWithClient(() => useSupportIdentifier("", true));

      expect(mockGetNamespaceSupportFn).not.toHaveBeenCalled();
    });

    it("does not fire the query when tenantId is null", async () => {
      const { useSupportIdentifier } = await importHook();

      renderHookWithClient(() => useSupportIdentifier(null, true));

      expect(mockGetNamespaceSupportFn).not.toHaveBeenCalled();
    });

    it("returns null identifier when disabled by empty tenantId", async () => {
      const { useSupportIdentifier } = await importHook();

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("", true),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe("when enabled with a valid tenant", () => {
    it("returns the identifier from the mocked response", async () => {
      mockGetNamespaceSupportFn.mockResolvedValue({ identifier: "abc123" });
      const { useSupportIdentifier } = await importHook();

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
      mockGetNamespaceSupportFn.mockRejectedValue(new Error("network error"));
      const { useSupportIdentifier } = await importHook();

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));

      // Initial attempt + 1 retry = 2 calls total. We don't retry further:
      // a misconfigured operator (4xx) shouldn't drag the spinner out.
      expect(mockGetNamespaceSupportFn).toHaveBeenCalledTimes(2);
    });
  });
});
