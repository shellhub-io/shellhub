import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useSupportIdentifier } from "../useSupportIdentifier";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespaceSupport: vi.fn(),
  }),
);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useSupportIdentifier", () => {
  describe("when enabled=false", () => {
    it("never fires the query and returns null identifier", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", false),
      );

      expect(sdk.getNamespaceSupport).not.toHaveBeenCalled();
      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("when tenantId is empty", () => {
    it("does not fire the query when tenantId is empty string", () => {
      renderHookWithClient(() => useSupportIdentifier("", true));

      expect(sdk.getNamespaceSupport).not.toHaveBeenCalled();
    });

    it("does not fire the query when tenantId is null", () => {
      renderHookWithClient(() => useSupportIdentifier(null, true));

      expect(sdk.getNamespaceSupport).not.toHaveBeenCalled();
    });

    it("returns null identifier when disabled by empty tenantId", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("", true),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe("when enabled with a valid tenant", () => {
    it("returns the identifier from the response", async () => {
      sdk.getNamespaceSupport.mockResolvedValue(
        mockSdkResponse({ identifier: "abc123" }),
      );

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.identifier).toBe("abc123"));
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("retry policy", () => {
    it("retries the query exactly once on failure (transient blip recovery)", async () => {
      sdk.getNamespaceSupport.mockRejectedValue(new Error("network error"));

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(sdk.getNamespaceSupport).toHaveBeenCalledTimes(2);
    });
  });
});
