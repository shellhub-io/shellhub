import { describe, it, expect, vi } from "vitest";
import { paginatedQueryFn, paginatedQueryOptions } from "../pagination";

function mockSdkFn(data: unknown[], headers: Record<string, string>) {
  return vi.fn().mockResolvedValue({
    data,
    response: { headers: new Headers(headers) },
  });
}

describe("paginatedQueryFn", () => {
  it("returns data and totalCount from X-Total-Count header", async () => {
    const devices = [{ uid: "1" }, { uid: "2" }];
    const sdkFn = mockSdkFn(devices, { "X-Total-Count": "42" });

    const queryFn = paginatedQueryFn(sdkFn, { query: { page: 1 } });
    const result = await queryFn();

    expect(result).toEqual({ data: devices, totalCount: 42 });
    expect(sdkFn).toHaveBeenCalledWith({ query: { page: 1 }, throwOnError: true });
  });

  it("defaults totalCount to 0 when header is missing", async () => {
    const sdkFn = mockSdkFn([], {});

    const result = await paginatedQueryFn(sdkFn, {})();

    expect(result.totalCount).toBe(0);
  });

  it("propagates SDK errors thrown with throwOnError", async () => {
    const sdkFn = vi.fn().mockRejectedValue(new Error("network failure"));

    await expect(paginatedQueryFn(sdkFn, {})()).rejects.toThrow("network failure");
  });
});

describe("paginatedQueryOptions", () => {
  it("keys the query on the options it fetches with", async () => {
    const sdkFn = mockSdkFn([{ uid: "1" }], { "X-Total-Count": "7" });
    const options = { query: { page: 2, per_page: 5 } };

    const { queryKey, queryFn } = paginatedQueryOptions(
      sdkFn,
      (o: typeof options) => ["getThings", o],
      options,
    );

    expect(queryKey).toEqual(["getThings", options]);
    expect(await queryFn()).toEqual({ data: [{ uid: "1" }], totalCount: 7 });
    expect(sdkFn).toHaveBeenCalledWith({ ...options, throwOnError: true });
  });
});
