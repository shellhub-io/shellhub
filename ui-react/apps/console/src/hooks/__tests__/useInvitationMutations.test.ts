import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAcceptInvite,
  useDeclineInvite,
  useGenerateInvitationLink,
  useCancelMembershipInvitation,
  useUpdateMembershipInvitation,
} from "../useInvitationMutations";

const mockAcceptFn = vi.fn();
const mockDeclineFn = vi.fn();
const mockGenerateLinkFn = vi.fn();
const mockCancelFn = vi.fn();
const mockUpdateFn = vi.fn();
const mockInvalidate = vi.fn();

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  acceptInviteMutation: vi.fn(() => ({ mutationFn: mockAcceptFn })),
  declineInviteMutation: vi.fn(() => ({ mutationFn: mockDeclineFn })),
  generateInvitationLinkMutation: vi.fn(() => ({
    mutationFn: mockGenerateLinkFn,
  })),
  cancelMembershipInvitationMutation: vi.fn(() => ({
    mutationFn: mockCancelFn,
  })),
  updateMembershipInvitationMutation: vi.fn(() => ({
    mutationFn: mockUpdateFn,
  })),
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

describe("useAcceptInvite", () => {
  describe("mutation call", () => {
    it("calls acceptInvite with the provided path", async () => {
      mockAcceptFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useAcceptInvite(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { tenant: "t1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAcceptFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockAcceptFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useAcceptInvite(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("accept failed");
      mockAcceptFn.mockRejectedValue(error);
      const { result } = renderHook(() => useAcceptInvite(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockAcceptFn.mockRejectedValue(new Error("accept failed"));
      const { result } = renderHook(() => useAcceptInvite(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useDeclineInvite", () => {
  describe("mutation call", () => {
    it("calls declineInvite with the provided path", async () => {
      mockDeclineFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useDeclineInvite(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { tenant: "t1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockDeclineFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockDeclineFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useDeclineInvite(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("decline failed");
      mockDeclineFn.mockRejectedValue(error);
      const { result } = renderHook(() => useDeclineInvite(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockDeclineFn.mockRejectedValue(new Error("decline failed"));
      const { result } = renderHook(() => useDeclineInvite(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useGenerateInvitationLink", () => {
  describe("mutation call", () => {
    it("calls generateInvitationLink with path and body", async () => {
      mockGenerateLinkFn.mockResolvedValue({ link: "https://example.com/invite/abc" });
      const { result } = renderHook(() => useGenerateInvitationLink(), {
        wrapper: createWrapper(),
      });

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
      const { result } = renderHook(() => useGenerateInvitationLink(), {
        wrapper: createWrapper(),
      });

      const data = await act(() => result.current.mutateAsync({} as never));

      expect(data).toEqual({ link });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockGenerateLinkFn.mockResolvedValue({ link: "https://example.com/invite/abc" });
      const { result } = renderHook(() => useGenerateInvitationLink(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("generate link failed");
      mockGenerateLinkFn.mockRejectedValue(error);
      const { result } = renderHook(() => useGenerateInvitationLink(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockGenerateLinkFn.mockRejectedValue(new Error("generate link failed"));
      const { result } = renderHook(() => useGenerateInvitationLink(), {
        wrapper: createWrapper(),
      });

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
      const { result } = renderHook(() => useCancelMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { tenant: "t1", "user-id": "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockCancelFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockCancelFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useCancelMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("cancel failed");
      mockCancelFn.mockRejectedValue(error);
      const { result } = renderHook(() => useCancelMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockCancelFn.mockRejectedValue(new Error("cancel failed"));
      const { result } = renderHook(() => useCancelMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useUpdateMembershipInvitation", () => {
  describe("mutation call", () => {
    it("calls updateMembershipInvitation with path and body", async () => {
      mockUpdateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUpdateMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      const vars = {
        path: { tenant: "t1", "user-id": "u1" },
        body: { role: "administrator" },
      };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockUpdateFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockUpdateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUpdateMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("update failed");
      mockUpdateFn.mockRejectedValue(error);
      const { result } = renderHook(() => useUpdateMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockUpdateFn.mockRejectedValue(new Error("update failed"));
      const { result } = renderHook(() => useUpdateMembershipInvitation(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
