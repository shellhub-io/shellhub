import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

afterEach(() => {
  cleanup();
});

// ── SDK mocks ────────────────────────────────────────────────────────────────

const mockMostUsedQueryFn = vi.fn();
const mockChoiceMutationFn = vi.fn();
const mockInvalidate = vi.fn();

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  getDevicesMostUsedOptions: vi.fn(() => ({
    queryKey: [{ _id: "getDevicesMostUsed" }],
    queryFn: mockMostUsedQueryFn,
  })),
  choiceDevicesMutation: vi.fn(() => ({
    mutationFn: mockChoiceMutationFn,
  })),
}));

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

// ── Imports under test (vi.mock calls hoist above these) ─────────────────────

import { useSuggestedDevices, useChoiceDevices } from "../useDeviceChooser";
import { useInvalidateByIds } from "../useInvalidateQueries";

// ── Helpers ──────────────────────────────────────────────────────────────────

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

// ── useSuggestedDevices ──────────────────────────────────────────────────────

describe("useSuggestedDevices", () => {
  describe("on success", () => {
    it("returns the device list from the query response", async () => {
      const devices = [
        { uid: "d1", name: "host-1" },
        { uid: "d2", name: "host-2" },
      ];
      mockMostUsedQueryFn.mockResolvedValue(devices);

      const { result } = renderHook(() => useSuggestedDevices(), {
        wrapper: createWrapper(),
      });

      await waitFor(() =>
        expect(result.current.devices).toEqual([
          { uid: "d1", name: "host-1", tags: [] },
          { uid: "d2", name: "host-2", tags: [] },
        ]),
      );
    });

    it("returns an empty array when data is undefined before load completes", () => {
      // Don't resolve yet — keep the query pending
      mockMostUsedQueryFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useSuggestedDevices(), {
        wrapper: createWrapper(),
      });

      expect(result.current.devices).toEqual([]);
    });

    it("exposes isLoading=true while the query is in flight", () => {
      mockMostUsedQueryFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useSuggestedDevices(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });
  });

  describe("when enabled=false", () => {
    it("does not call the queryFn", () => {
      renderHook(() => useSuggestedDevices(false), {
        wrapper: createWrapper(),
      });

      expect(mockMostUsedQueryFn).not.toHaveBeenCalled();
    });

    it("returns an empty devices array", () => {
      const { result } = renderHook(() => useSuggestedDevices(false), {
        wrapper: createWrapper(),
      });

      expect(result.current.devices).toEqual([]);
    });
  });

  describe("on error", () => {
    it("exposes the error on failure", async () => {
      const err = new Error("network failure");
      mockMostUsedQueryFn.mockRejectedValue(err);

      const { result } = renderHook(() => useSuggestedDevices(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.error).toBe(err));
    });
  });
});

// ── useChoiceDevices ─────────────────────────────────────────────────────────

describe("useChoiceDevices", () => {
  describe("mutation call", () => {
    it("calls the mutationFn with the choices body", async () => {
      mockChoiceMutationFn.mockResolvedValue(undefined);

      const { result } = renderHook(() => useChoiceDevices(), {
        wrapper: createWrapper(),
      });

      const vars = { body: { choices: ["uid1", "uid2"] } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockChoiceMutationFn).toHaveBeenCalledWith(
        vars,
        expect.anything(),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate once after the mutation succeeds", async () => {
      mockChoiceMutationFn.mockResolvedValue(undefined);

      const { result } = renderHook(() => useChoiceDevices(), {
        wrapper: createWrapper(),
      });

      await act(() =>
        result.current.mutateAsync({ body: { choices: ["uid1"] } }),
      );

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });

    it("registers invalidate using useInvalidateByIds with the correct query ids", () => {
      renderHook(() => useChoiceDevices(), { wrapper: createWrapper() });

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
      mockChoiceMutationFn.mockRejectedValue(err);

      const { result } = renderHook(() => useChoiceDevices(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(err);
    });

    it("does not call invalidate when the mutation fails", async () => {
      mockChoiceMutationFn.mockRejectedValue(new Error("server error"));

      const { result } = renderHook(() => useChoiceDevices(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({ body: { choices: ["uid1"] } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
