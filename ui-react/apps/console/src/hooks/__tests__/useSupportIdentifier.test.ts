import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const mockGetNamespaceSupportFn = vi.fn();

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getNamespaceSupportOptions: vi.fn(() => ({
    queryKey: [{ _id: "getNamespaceSupport" }],
    queryFn: mockGetNamespaceSupportFn,
  })),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

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

      const { result } = renderHook(
        () => useSupportIdentifier("tenant-123", false),
        { wrapper: createWrapper() },
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

      renderHook(() => useSupportIdentifier("", true), {
        wrapper: createWrapper(),
      });

      expect(mockGetNamespaceSupportFn).not.toHaveBeenCalled();
    });

    it("does not fire the query when tenantId is null", async () => {
      const { useSupportIdentifier } = await importHook();

      renderHook(() => useSupportIdentifier(null, true), {
        wrapper: createWrapper(),
      });

      expect(mockGetNamespaceSupportFn).not.toHaveBeenCalled();
    });

    it("returns null identifier when disabled by empty tenantId", async () => {
      const { useSupportIdentifier } = await importHook();

      const { result } = renderHook(() => useSupportIdentifier("", true), {
        wrapper: createWrapper(),
      });

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe("when enabled with a valid tenant", () => {
    it("returns the identifier from the mocked response", async () => {
      mockGetNamespaceSupportFn.mockResolvedValue({ identifier: "abc123" });
      const { useSupportIdentifier } = await importHook();

      const { result } = renderHook(
        () => useSupportIdentifier("tenant-123", true),
        { wrapper: createWrapper() },
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

      const { result } = renderHook(
        () => useSupportIdentifier("tenant-123", true),
        { wrapper: createWrapper() },
      );

      await waitFor(() => expect(result.current.isError).toBe(true));

      // Initial attempt + 1 retry = 2 calls total. We don't retry further:
      // a misconfigured operator (4xx) shouldn't drag the spinner out.
      expect(mockGetNamespaceSupportFn).toHaveBeenCalledTimes(2);
    });
  });
});
