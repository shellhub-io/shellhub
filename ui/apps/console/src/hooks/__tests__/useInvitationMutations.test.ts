import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  useAcceptInvite,
  useGenerateInvitationLink,
  useCancelMembershipInvitation,
} from "../useInvitationMutations";

const mockInvalidate = vi.fn();

const sdk = vi.hoisted(() =>
  mockSdkGen({
    acceptInvite: vi.fn(),
    generateInvitationLink: vi.fn(),
    cancelMembershipInvitation: vi.fn(),
  }),
);

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useAcceptInvite", () => {
  describe("mutation call", () => {
    it("calls acceptInvite with the provided path", async () => {
      sdk.acceptInvite.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      const vars = { path: { tenant: "t1" } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.acceptInvite).toHaveBeenCalledWith(
        expect.objectContaining({ path: { tenant: "t1" }, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      sdk.acceptInvite.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      await act(() => result.current.mutateAsync({ path: { tenant: "t1" } }));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("accept failed");
      sdk.acceptInvite.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useAcceptInvite());

      act(() => result.current.mutate({ path: { tenant: "t1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      sdk.acceptInvite.mockRejectedValue(new Error("accept failed"));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      act(() => result.current.mutate({ path: { tenant: "t1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useGenerateInvitationLink", () => {
  describe("mutation call", () => {
    it("calls generateInvitationLink with path and body", async () => {
      sdk.generateInvitationLink.mockResolvedValue(
        mockSdkResponse({ link: "https://example.com/invite/abc" }),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const vars = {
        path: { tenant: "t1" },
        body: { email: "bob@example.com", role: "operator" as const },
      };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.generateInvitationLink).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { tenant: "t1" },
          body: { email: "bob@example.com", role: "operator" },
          throwOnError: true,
        }),
      );
    });

    it("returns the generated link from the mutation", async () => {
      const link = "https://example.com/invite/xyz";
      sdk.generateInvitationLink.mockResolvedValue(mockSdkResponse({ link }));
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const data = await act(() =>
        result.current.mutateAsync({ path: { tenant: "t1" } }),
      );

      expect(data).toEqual({ link });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      sdk.generateInvitationLink.mockResolvedValue(
        mockSdkResponse({ link: "https://example.com/invite/abc" }),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      await act(() => result.current.mutateAsync({ path: { tenant: "t1" } }));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("generate link failed");
      sdk.generateInvitationLink.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      act(() => result.current.mutate({ path: { tenant: "t1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      sdk.generateInvitationLink.mockRejectedValue(
        new Error("generate link failed"),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      act(() => result.current.mutate({ path: { tenant: "t1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useCancelMembershipInvitation", () => {
  describe("mutation call", () => {
    it("calls cancelMembershipInvitation with path", async () => {
      sdk.cancelMembershipInvitation.mockResolvedValue(
        mockSdkResponse(undefined),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      const vars = { path: { tenant: "t1", "user-id": "u1" } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.cancelMembershipInvitation).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { tenant: "t1", "user-id": "u1" },
          throwOnError: true,
        }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      sdk.cancelMembershipInvitation.mockResolvedValue(
        mockSdkResponse(undefined),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      await act(() =>
        result.current.mutateAsync({
          path: { tenant: "t1", "user-id": "u1" },
        }),
      );

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("cancel failed");
      sdk.cancelMembershipInvitation.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() =>
        result.current.mutate({
          path: { tenant: "t1", "user-id": "u1" },
        }),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      sdk.cancelMembershipInvitation.mockRejectedValue(
        new Error("cancel failed"),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() =>
        result.current.mutate({
          path: { tenant: "t1", "user-id": "u1" },
        }),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
