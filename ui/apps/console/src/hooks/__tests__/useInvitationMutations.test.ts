import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  useAcceptInvite,
  useGenerateInvitationLink,
  useCancelMembershipInvitation,
} from "../useInvitationMutations";

const mockAcceptInvite = vi.hoisted(() => vi.fn());
const mockGenerateInvitationLink = vi.hoisted(() => vi.fn());
const mockCancelMembershipInvitation = vi.hoisted(() => vi.fn());
const mockInvalidate = vi.fn();

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    acceptInvite: mockAcceptInvite,
    generateInvitationLink: mockGenerateInvitationLink,
    cancelMembershipInvitation: mockCancelMembershipInvitation,
  };
});

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useAcceptInvite", () => {
  describe("mutation call", () => {
    it("calls acceptInvite with the provided path", async () => {
      mockAcceptInvite.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      const vars = { path: { tenant: "t1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAcceptInvite).toHaveBeenCalledWith(
        expect.objectContaining({ path: { tenant: "t1" }, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockAcceptInvite.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("accept failed");
      mockAcceptInvite.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useAcceptInvite());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockAcceptInvite.mockRejectedValue(new Error("accept failed"));
      const { result } = renderHookWithClient(() => useAcceptInvite());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useGenerateInvitationLink", () => {
  describe("mutation call", () => {
    it("calls generateInvitationLink with path and body", async () => {
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({ link: "https://example.com/invite/abc" }),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const vars = {
        path: { tenant: "t1" },
        body: { email: "bob@example.com", role: "operator" },
      };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockGenerateInvitationLink).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { tenant: "t1" },
          body: { email: "bob@example.com", role: "operator" },
          throwOnError: true,
        }),
      );
    });

    it("returns the generated link from the mutation", async () => {
      const link = "https://example.com/invite/xyz";
      mockGenerateInvitationLink.mockResolvedValue(mockSdkResponse({ link }));
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const data = await act(() => result.current.mutateAsync({} as never));

      expect(data).toEqual({ link });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({ link: "https://example.com/invite/abc" }),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("generate link failed");
      mockGenerateInvitationLink.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockGenerateInvitationLink.mockRejectedValue(
        new Error("generate link failed"),
      );
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useCancelMembershipInvitation", () => {
  describe("mutation call", () => {
    it("calls cancelMembershipInvitation with path", async () => {
      mockCancelMembershipInvitation.mockResolvedValue(
        mockSdkResponse(undefined),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      const vars = { path: { tenant: "t1", "user-id": "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockCancelMembershipInvitation).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { tenant: "t1", "user-id": "u1" },
          throwOnError: true,
        }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockCancelMembershipInvitation.mockResolvedValue(
        mockSdkResponse(undefined),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("cancel failed");
      mockCancelMembershipInvitation.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockCancelMembershipInvitation.mockRejectedValue(
        new Error("cancel failed"),
      );
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
