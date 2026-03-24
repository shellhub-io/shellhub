import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act, cleanup } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useUploadLicense } from "../useUploadLicense";

afterEach(() => { cleanup(); });

const mockMutationFn = vi.fn();
const mockInvalidate = vi.fn();

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  sendLicenseMutation: vi.fn(() => ({ mutationFn: mockMutationFn })),
}));

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useUploadLicense", () => {
  describe("mutation call", () => {
    it("calls sendLicense with the provided body", async () => {
      mockMutationFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUploadLicense(), {
        wrapper: createWrapper(),
      });

      const body = { body: "license-data" };
      await act(() => result.current.mutateAsync(body as never));

      expect(mockMutationFn).toHaveBeenCalledWith(body, expect.anything());
    });
  });

  describe("on success", () => {
    it("invalidates getLicense queries after a successful mutation", async () => {
      mockMutationFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUploadLicense(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("rejects and exposes the error when sendLicense fails", async () => {
      const error = new Error("upload failed");
      mockMutationFn.mockRejectedValue(error);
      const { result } = renderHook(() => useUploadLicense(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when sendLicense fails", async () => {
      mockMutationFn.mockRejectedValue(new Error("upload failed"));
      const { result } = renderHook(() => useUploadLicense(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
