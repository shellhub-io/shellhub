import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { usePublicKeys } from "../usePublicKeys";
import type { PublicKeyResponse } from "@/client";

const mockGetPublicKeys = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getPublicKeys: mockGetPublicKeys };
});

function makeKey(
  overrides: Partial<PublicKeyResponse> = {},
): PublicKeyResponse {
  return {
    name: "test-key",
    fingerprint: "aa:bb:cc",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-1",
    data: "c3NoLXJzYQ==",
    filter: { hostname: ".*", tags: [] },
    username: ".*",
    ...overrides,
  };
}

function makeTag(name: string) {
  return {
    name,
    tenant_id: "tenant-1",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("usePublicKeys", () => {
  describe("returns", () => {
    it("returns publicKeys from the paginated result", async () => {
      const keys = [
        makeKey({ name: "key-1", filter: { hostname: ".*", tags: [] } }),
        makeKey({ name: "key-2", filter: { tags: [makeTag("prod")] } }),
      ];
      mockGetPublicKeys.mockResolvedValue(
        mockSdkResponse(keys, { "X-Total-Count": "2" }),
      );

      const { result } = renderHookWithClient(() => usePublicKeys());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.publicKeys).toHaveLength(2);
      expect(result.current.publicKeys[0].name).toBe("key-1");
      expect(result.current.publicKeys[1].name).toBe("key-2");
    });

    it("returns totalCount from the X-Total-Count header", async () => {
      mockGetPublicKeys.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "42" }),
      );

      const { result } = renderHookWithClient(() => usePublicKeys());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(42);
    });

    it("defaults publicKeys to empty array while loading", () => {
      mockGetPublicKeys.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => usePublicKeys());

      expect(result.current.publicKeys).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetPublicKeys.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => usePublicKeys());

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetPublicKeys.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => usePublicKeys());

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetPublicKeys.mockRejectedValue(networkError);

      const { result } = renderHookWithClient(() => usePublicKeys());

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetPublicKeys.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => usePublicKeys());

      await waitFor(() => expect(mockGetPublicKeys).toHaveBeenCalled());
      const [opts] = mockGetPublicKeys.mock.calls[0];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetPublicKeys.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => usePublicKeys({ page: 3, perPage: 25 }));

      await waitFor(() => expect(mockGetPublicKeys).toHaveBeenCalled());
      const [opts] = mockGetPublicKeys.mock.calls[0];
      expect(opts.query.page).toBe(3);
      expect(opts.query.per_page).toBe(25);
    });
  });
});
