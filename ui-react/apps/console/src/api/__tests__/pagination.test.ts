import { describe, it, expect, vi } from "vitest";
import { paginatedQueryFn } from "../pagination";

function mockSdkFn(data: unknown[] | undefined, headers: Record<string, string>, error?: unknown) {
  return vi.fn().mockResolvedValue({
    data,
    error,
    response: {
      headers: new Headers(headers),
    },
  });
}

describe("paginatedQueryFn", () => {
  it("returns data and totalCount from X-Total-Count header", async () => {
    const devices = [{ uid: "1" }, { uid: "2" }];
    const sdkFn = mockSdkFn(devices, { "X-Total-Count": "42" });

    const queryFn = paginatedQueryFn(sdkFn, { query: { page: 1 } });
    const result = await queryFn();

    expect(result).toEqual({ data: devices, totalCount: 42 });
    expect(sdkFn).toHaveBeenCalledWith({ query: { page: 1 } });
  });

  it("defaults totalCount to 0 when header is missing", async () => {
    const sdkFn = mockSdkFn([], {});

    const result = await paginatedQueryFn(sdkFn, {})();

    expect(result.totalCount).toBe(0);
  });

  it("defaults data to empty array when undefined", async () => {
    const sdkFn = mockSdkFn(undefined, { "X-Total-Count": "0" });

    const result = await paginatedQueryFn(sdkFn, {})();

    expect(result.data).toEqual([]);
  });

  it("throws when error is present", async () => {
    const sdkFn = mockSdkFn(undefined, {}, { message: "not found" });

    await expect(paginatedQueryFn(sdkFn, {})()).rejects.toThrow();
  });

  it("throws Error instances directly", async () => {
    const originalError = new Error("network failure");
    const sdkFn = mockSdkFn(undefined, {}, originalError);

    await expect(paginatedQueryFn(sdkFn, {})()).rejects.toBe(originalError);
  });
});
