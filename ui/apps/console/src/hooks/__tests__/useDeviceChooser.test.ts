import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useSuggestedDevices, useChoiceDevices } from "../useDeviceChooser";
import { useInvalidateByIds } from "../useInvalidateQueries";

const mockInvalidate = vi.fn();

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getDevicesMostUsed: vi.fn(),
    choiceDevices: vi.fn(),
  }),
);

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
      sdk.getDevicesMostUsed.mockResolvedValue(mockSdkResponse(devices));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      await waitFor(() =>
        expect(result.current.devices).toEqual([
          { uid: "d1", name: "host-1", tags: [] },
          { uid: "d2", name: "host-2", tags: [] },
        ]),
      );
    });

    it("returns an empty array when data is undefined before load completes", () => {
      sdk.getDevicesMostUsed.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      expect(result.current.devices).toEqual([]);
    });

    it("exposes isLoading=true while the query is in flight", () => {
      sdk.getDevicesMostUsed.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      expect(result.current.isLoading).toBe(true);
    });
  });

  describe("when enabled=false", () => {
    it("does not call the SDK function", () => {
      renderHookWithClient(() => useSuggestedDevices(false));

      expect(sdk.getDevicesMostUsed).not.toHaveBeenCalled();
    });

    it("returns an empty devices array", () => {
      const { result } = renderHookWithClient(() => useSuggestedDevices(false));

      expect(result.current.devices).toEqual([]);
    });
  });

  describe("on error", () => {
    it("exposes the error on failure", async () => {
      const err = new Error("network failure");
      sdk.getDevicesMostUsed.mockRejectedValue(err);

      const { result } = renderHookWithClient(() => useSuggestedDevices());

      await waitFor(() => expect(result.current.error).toBe(err));
    });
  });
});

describe("useChoiceDevices", () => {
  describe("mutation call", () => {
    it("calls the SDK function with the choices body", async () => {
      sdk.choiceDevices.mockResolvedValue(mockSdkResponse(undefined));

      const { result } = renderHookWithClient(() => useChoiceDevices());

      const vars = { body: { choices: ["uid1", "uid2"] } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.choiceDevices).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { choices: ["uid1", "uid2"] },
          throwOnError: true,
        }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate once after the mutation succeeds", async () => {
      sdk.choiceDevices.mockResolvedValue(mockSdkResponse(undefined));

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
      sdk.choiceDevices.mockRejectedValue(err);

      const { result } = renderHookWithClient(() => useChoiceDevices());

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(err);
    });

    it("does not call invalidate when the mutation fails", async () => {
      sdk.choiceDevices.mockRejectedValue(new Error("server error"));

      const { result } = renderHookWithClient(() => useChoiceDevices());

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
