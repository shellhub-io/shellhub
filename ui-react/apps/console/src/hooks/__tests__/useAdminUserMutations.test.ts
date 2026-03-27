import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  useResetUserPassword,
} from "../useAdminUserMutations";

const mockCreateFn = vi.fn();
const mockUpdateFn = vi.fn();
const mockDeleteFn = vi.fn();
const mockResetPasswordFn = vi.fn();
const mockInvalidate = vi.fn();

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  createUserAdminMutation: vi.fn(() => ({ mutationFn: mockCreateFn })),
  adminUpdateUserMutation: vi.fn(() => ({ mutationFn: mockUpdateFn })),
  adminDeleteUserMutation: vi.fn(() => ({ mutationFn: mockDeleteFn })),
  adminResetUserPasswordMutation: vi.fn(() => ({
    mutationFn: mockResetPasswordFn,
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

describe("useCreateUser", () => {
  describe("mutation call", () => {
    it("calls createUserAdmin with the provided body", async () => {
      mockCreateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useCreateUser(), {
        wrapper: createWrapper(),
      });

      const body = {
        body: {
          name: "Alice",
          username: "alice",
          email: "alice@example.com",
          password: "pass1",
        },
      };
      await act(() => result.current.mutateAsync(body as never));

      expect(mockCreateFn).toHaveBeenCalledWith(body, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockCreateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useCreateUser(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("create failed");
      mockCreateFn.mockRejectedValue(error);
      const { result } = renderHook(() => useCreateUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockCreateFn.mockRejectedValue(new Error("create failed"));
      const { result } = renderHook(() => useCreateUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useUpdateUser", () => {
  describe("mutation call", () => {
    it("calls adminUpdateUser with path and body", async () => {
      mockUpdateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { id: "u1" }, body: { name: "Bob" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockUpdateFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful update", async () => {
      mockUpdateFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when update fails", async () => {
      const error = new Error("update failed");
      mockUpdateFn.mockRejectedValue(error);
      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when update fails", async () => {
      mockUpdateFn.mockRejectedValue(new Error("update failed"));
      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useDeleteUser", () => {
  describe("mutation call", () => {
    it("calls adminDeleteUser with the path", async () => {
      mockDeleteFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useDeleteUser(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockDeleteFn).toHaveBeenCalledWith(vars, expect.anything());
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful delete", async () => {
      mockDeleteFn.mockResolvedValue(undefined);
      const { result } = renderHook(() => useDeleteUser(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when delete fails", async () => {
      const error = new Error("delete failed");
      mockDeleteFn.mockRejectedValue(error);
      const { result } = renderHook(() => useDeleteUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when delete fails", async () => {
      mockDeleteFn.mockRejectedValue(new Error("delete failed"));
      const { result } = renderHook(() => useDeleteUser(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useResetUserPassword", () => {
  describe("mutation call", () => {
    it("calls adminResetUserPassword with the path", async () => {
      mockResetPasswordFn.mockResolvedValue({ password: "generated-pw" });
      const { result } = renderHook(() => useResetUserPassword(), {
        wrapper: createWrapper(),
      });

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockResetPasswordFn).toHaveBeenCalledWith(vars, expect.anything());
    });

    it("returns the generated password from the mutation", async () => {
      mockResetPasswordFn.mockResolvedValue({ password: "s3cr3t-pass" });
      const { result } = renderHook(() => useResetUserPassword(), {
        wrapper: createWrapper(),
      });

      const data = await act(() => result.current.mutateAsync({} as never));

      expect(data).toEqual({ password: "s3cr3t-pass" });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful password reset", async () => {
      mockResetPasswordFn.mockResolvedValue({ password: "pw" });
      const { result } = renderHook(() => useResetUserPassword(), {
        wrapper: createWrapper(),
      });

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when reset fails", async () => {
      const error = new Error("reset failed");
      mockResetPasswordFn.mockRejectedValue(error);
      const { result } = renderHook(() => useResetUserPassword(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when reset fails", async () => {
      mockResetPasswordFn.mockRejectedValue(new Error("reset failed"));
      const { result } = renderHook(() => useResetUserPassword(), {
        wrapper: createWrapper(),
      });

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
