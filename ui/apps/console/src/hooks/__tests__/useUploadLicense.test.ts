import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useUploadLicense } from "../useUploadLicense";

const mockSendLicense = vi.hoisted(() => vi.fn());
const mockInvalidate = vi.fn();

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, sendLicense: mockSendLicense };
});

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useUploadLicense", () => {
  describe("mutation call", () => {
    it("calls sendLicense with the provided body", async () => {
      mockSendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUploadLicense());

      const body = { body: "license-data" };
      await act(() => result.current.mutateAsync(body as never));

      expect(mockSendLicense).toHaveBeenCalledWith(
        expect.objectContaining({ body: "license-data", throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("invalidates getLicense queries after a successful mutation", async () => {
      mockSendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUploadLicense());

      await act(() => result.current.mutateAsync({}));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("rejects and exposes the error when sendLicense fails", async () => {
      const error = new Error("upload failed");
      mockSendLicense.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useUploadLicense());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when sendLicense fails", async () => {
      mockSendLicense.mockRejectedValue(new Error("upload failed"));
      const { result } = renderHookWithClient(() => useUploadLicense());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
