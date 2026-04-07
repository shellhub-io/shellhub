import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { usePublicKeys } from "../usePublicKeys";
import type { PublicKeyResponse } from "../../client";

vi.mock("../../client", () => ({
  getPublicKeys: vi.fn(),
}));

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  getPublicKeysQueryKey: vi.fn((opts: unknown) => [{ _id: "getPublicKeys" }, opts]),
}));

vi.mock("../../api/pagination", () => ({
  paginatedQueryFn: vi.fn(
    (_sdkFn: unknown, opts: { query: Record<string, unknown> }) => {
      return () => mockGetPublicKeysFn(opts) as unknown;
    },
  ),
}));

const mockGetPublicKeysFn = vi.fn();

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function makeKey(overrides: Partial<PublicKeyResponse> = {}): PublicKeyResponse {
  return {
    name: "test-key",
    fingerprint: "aa:bb:cc",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-1",
    data: "c3NoLXJzYQ==",
    filter: { hostname: ".*" },
    username: ".*",
    ...overrides,
  } as PublicKeyResponse;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("usePublicKeys", () => {
  describe("normalizePublicKey — hostname filter", () => {
    it("sets filter.hostname from response", async () => {
      const key = makeKey({ filter: { hostname: "^prod-.*" } });
      mockGetPublicKeysFn.mockResolvedValue({ data: [key], totalCount: 1 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys[0].filter).toEqual({ hostname: "^prod-.*" });
    });

    it("sets filter with hostname '.*' for the catch-all case", async () => {
      const key = makeKey({ filter: { hostname: ".*" } });
      mockGetPublicKeysFn.mockResolvedValue({ data: [key], totalCount: 1 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys[0].filter).toEqual({ hostname: ".*" });
    });
  });

  describe("normalizePublicKey — tags filter", () => {
    it("sets filter.tags as string array directly from response", async () => {
      const key = makeKey({ filter: { tags: ["production", "linux"] } });
      mockGetPublicKeysFn.mockResolvedValue({ data: [key], totalCount: 1 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys[0].filter).toEqual({
        tags: ["production", "linux"],
      });
    });

    it("preserves a single-element tags array", async () => {
      const key = makeKey({ filter: { tags: ["web"] } });
      mockGetPublicKeysFn.mockResolvedValue({ data: [key], totalCount: 1 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys[0].filter.tags).toEqual(["web"]);
    });

    it("does not set filter.hostname when tags are present", async () => {
      const key = makeKey({ filter: { tags: ["api"] } });
      mockGetPublicKeysFn.mockResolvedValue({ data: [key], totalCount: 1 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys[0].filter.hostname).toBeUndefined();
    });
  });

  describe("returns", () => {
    it("returns publicKeys from the paginated result", async () => {
      const keys = [
        makeKey({ name: "key-1", filter: { hostname: ".*" } }),
        makeKey({ name: "key-2", filter: { tags: ["prod"] } }),
      ];
      mockGetPublicKeysFn.mockResolvedValue({ data: keys, totalCount: 2 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys).toHaveLength(2);
      expect(result.current.publicKeys[0].name).toBe("key-1");
      expect(result.current.publicKeys[1].name).toBe("key-2");
    });

    it("returns totalCount from the paginated result", async () => {
      mockGetPublicKeysFn.mockResolvedValue({ data: [], totalCount: 42 });

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(42);
    });

    it("defaults publicKeys to empty array while loading", () => {
      mockGetPublicKeysFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      expect(result.current.publicKeys).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetPublicKeysFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetPublicKeysFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetPublicKeysFn.mockRejectedValue(networkError);

      const { result } = renderHook(() => usePublicKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetPublicKeysFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => usePublicKeys(), { wrapper: createWrapper() });

      await waitFor(() => expect(mockGetPublicKeysFn).toHaveBeenCalled());
      const [opts] = mockGetPublicKeysFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetPublicKeysFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => usePublicKeys({ page: 3, perPage: 25 }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockGetPublicKeysFn).toHaveBeenCalled());
      const [opts] = mockGetPublicKeysFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(3);
      expect(opts.query.per_page).toBe(25);
    });
  });
});
