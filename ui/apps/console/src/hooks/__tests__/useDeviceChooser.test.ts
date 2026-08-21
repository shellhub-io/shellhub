import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useSuggestedDevices, useChoiceDevices } from "../useDeviceChooser";
import { useInvalidateByIds } from "../useInvalidateQueries";

const mockGetDevicesMostUsed = vi.hoisted(() => vi.fn());
const mockChoiceDevices = vi.hoisted(() => vi.fn());
const mockInvalidate = vi.fn();

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    getDevicesMostUsed: mockGetDevicesMostUsed,
    choiceDevices: mockChoiceDevices,
  };
});

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useSuggestedDevices", () => {
  describe("on success", () => {
    it("returns the device list from the query response", async () => {
      const devices = [
        { uid: "d1", name: "host-1" },
        { uid: "d2", name: "host-2" },
      ];
      mockGetDevicesMostUsed.mockResolvedValue(mockSdkResponse(devices));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      await waitFor(() =>
        expect(result.current.devices).toEqual([
          { uid: "d1", name: "host-1", tags: [] },
          { uid: "d2", name: "host-2", tags: [] },
        ]),
      );
    });

    it("returns an empty array when data is undefined before load completes", () => {
      mockGetDevicesMostUsed.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      expect(result.current.devices).toEqual([]);
    });

    it("exposes isLoading=true while the query is in flight", () => {
      mockGetDevicesMostUsed.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      expect(result.current.isLoading).toBe(true);
    });
  });

  describe("when enabled=false", () => {
    it("does not call the SDK function", () => {
      renderHookWithClient(() => useSuggestedDevices(false));

      expect(mockGetDevicesMostUsed).not.toHaveBeenCalled();
    });

    it("returns an empty devices array", () => {
      const { result } = renderHookWithClient(() => useSuggestedDevices(false));

      expect(result.current.devices).toEqual([]);
    });
  });

  describe("on error", () => {
    it("exposes the error on failure", async () => {
      const err = new Error("network failure");
      mockGetDevicesMostUsed.mockRejectedValue(err);

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      await waitFor(() => expect(result.current.error).toBe(err));
    });
  });
});

describe("useChoiceDevices", () => {
  describe("mutation call", () => {
    it("calls the SDK function with the choices body", async () => {
      mockChoiceDevices.mockResolvedValue(mockSdkResponse(undefined));

      const { result } = renderHookWithClient(() => useChoiceDevices());

      const vars = { body: { choices: ["uid1", "uid2"] } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockChoiceDevices).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { choices: ["uid1", "uid2"] },
          throwOnError: true,
        }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate once after the mutation succeeds", async () => {
      mockChoiceDevices.mockResolvedValue(mockSdkResponse(undefined));

      const { result } = renderHookWithClient(() => useChoiceDevices());

      await act(() =>
        result.current.mutateAsync({ body: { choices: ["uid1"] } }),
      );

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });

    it("registers invalidate using useInvalidateByIds with the correct query ids", () => {
      renderHookWithClient(() => useChoiceDevices());

      expect(useInvalidateByIds).toHaveBeenCalledWith(
        "getDevices",
        "getDevice",
        "getStatusDevices",
      );
    });
  });

  describe("on failure", () => {
    it("exposes error when the mutation fails", async () => {
      const err = new Error("server error");
      mockChoiceDevices.mockRejectedValue(err);

      const { result } = renderHookWithClient(() => useChoiceDevices());

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(err);
    });

    it("does not call invalidate when the mutation fails", async () => {
      mockChoiceDevices.mockRejectedValue(new Error("server error"));

      const { result } = renderHookWithClient(() => useChoiceDevices());

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
