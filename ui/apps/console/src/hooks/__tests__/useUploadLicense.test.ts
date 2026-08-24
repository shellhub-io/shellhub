import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useUploadLicense } from "../useUploadLicense";

const mockInvalidate = vi.fn();

const sdk = vi.hoisted(() =>
  mockSdkGen({
    sendLicense: vi.fn(),
  }),
);

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useUploadLicense", () => {
  describe("mutation call", () => {
    it("calls sendLicense with the provided body", async () => {
      sdk.sendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUploadLicense());

      const file = new File(["license-data"], "license.lic");
      await act(() => result.current.mutateAsync({ body: { file } }));

      expect(sdk.sendLicense).toHaveBeenCalledWith(
        expect.objectContaining({ body: { file }, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("invalidates getLicense queries after a successful mutation", async () => {
      sdk.sendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUploadLicense());

      await act(() => result.current.mutateAsync({}));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("rejects and exposes the error when sendLicense fails", async () => {
      const error = new Error("upload failed");
      sdk.sendLicense.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useUploadLicense());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when sendLicense fails", async () => {
      sdk.sendLicense.mockRejectedValue(new Error("upload failed"));
      const { result } = renderHookWithClient(() => useUploadLicense());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
