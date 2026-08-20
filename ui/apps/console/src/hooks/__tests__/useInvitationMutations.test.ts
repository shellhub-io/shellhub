import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import {
  useAcceptInvite,
  useGenerateInvitationLink,
  useCancelMembershipInvitation,
} from "../useInvitationMutations";

const mockAcceptFn = vi.fn();
const mockGenerateLinkFn = vi.fn();
const mockCancelFn = vi.fn();
const mockInvalidate = vi.fn();

vi.mock("@/client", () => ({
  acceptInviteMutation: vi.fn(() => ({ mutationFn: mockAcceptFn })),
  generateInvitationLinkMutation: vi.fn(() => ({
    mutationFn: mockGenerateLinkFn,
  })),
  cancelMembershipInvitationMutation: vi.fn(() => ({
    mutationFn: mockCancelFn,
  })),
}));

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useAcceptInvite", () => {
  describe("mutation call", () => {
    it("calls acceptInvite with the provided path", async () => {
      mockAcceptFn.mockResolvedValue(undefined);
      const { result } = renderHookWithClient(() => useAcceptInvite());

      const vars = { path: { tenant: "t1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAcceptFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockAcceptFn.mockResolvedValue(undefined);
      const { result } = renderHookWithClient(() => useAcceptInvite());

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("accept failed");
      mockAcceptFn.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useAcceptInvite());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockAcceptFn.mockRejectedValue(new Error("accept failed"));
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
      mockGenerateLinkFn.mockResolvedValue({
        link: "https://example.com/invite/abc",
      });
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const vars = {
        path: { tenant: "t1" },
        body: { email: "bob@example.com", role: "operator" },
      };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockGenerateLinkFn).toHaveBeenCalledWith(vars, expect.anything());
    });

    it("returns the generated link from the mutation", async () => {
      const link = "https://example.com/invite/xyz";
      mockGenerateLinkFn.mockResolvedValue({ link });
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      const data = await act(() => result.current.mutateAsync({} as never));

      expect(data).toEqual({ link });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockGenerateLinkFn.mockResolvedValue({
        link: "https://example.com/invite/abc",
      });
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
      mockGenerateLinkFn.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useGenerateInvitationLink(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockGenerateLinkFn.mockRejectedValue(new Error("generate link failed"));
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
      mockCancelFn.mockResolvedValue(undefined);
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      const vars = { path: { tenant: "t1", "user-id": "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockCancelFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockCancelFn.mockResolvedValue(undefined);
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
      mockCancelFn.mockRejectedValue(error);
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockCancelFn.mockRejectedValue(new Error("cancel failed"));
      const { result } = renderHookWithClient(() =>
        useCancelMembershipInvitation(),
      );

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
