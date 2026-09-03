import { describe, it, expect, vi } from "vitest";
import { paginatedQueryFn } from "../pagination";
import * as customInstanceModule from "../customInstance";

vi.mock("../customInstance", () => ({
  fetchWithHeaders: vi.fn(),
}));

const mockFetchWithHeaders = vi.mocked(customInstanceModule.fetchWithHeaders);

describe("paginatedQueryFn", () => {
  it("returns data and totalCount from X-Total-Count header", async () => {
    const devices = [{ uid: "1" }, { uid: "2" }];
    mockFetchWithHeaders.mockResolvedValue({
      data: devices,
      headers: new Headers({ "X-Total-Count": "42" }),
    });

    const queryFn = paginatedQueryFn("/api/devices");
    const result = await queryFn({ signal: AbortSignal.abort() });

    expect(result).toEqual({ data: devices, totalCount: 42 });
    expect(mockFetchWithHeaders).toHaveBeenCalledWith("/api/devices", {
      method: "GET",
      signal: expect.any(AbortSignal),
    });
  });

  it("defaults totalCount to 0 when header is missing", async () => {
    mockFetchWithHeaders.mockResolvedValue({
      data: [],
      headers: new Headers(),
    });

    const result = await paginatedQueryFn("/api/devices")({
      signal: AbortSignal.abort(),
    });

    expect(result.totalCount).toBe(0);
  });

  it("propagates fetch errors", async () => {
    mockFetchWithHeaders.mockRejectedValue(new Error("network failure"));

    await expect(
      paginatedQueryFn("/api/devices")({ signal: AbortSignal.abort() }),
    ).rejects.toThrow("network failure");
  });
});
