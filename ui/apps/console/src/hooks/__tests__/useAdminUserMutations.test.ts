import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor, act } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  useCreateUser,
  useUpdateUser,
  useDeleteUser,
  useResetUserPassword,
} from "../useAdminUserMutations";

const mockCreateUserAdmin = vi.hoisted(() => vi.fn());
const mockAdminUpdateUser = vi.hoisted(() => vi.fn());
const mockAdminDeleteUser = vi.hoisted(() => vi.fn());
const mockAdminResetUserPassword = vi.hoisted(() => vi.fn());
const mockInvalidate = vi.fn();

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    createUserAdmin: mockCreateUserAdmin,
    adminUpdateUser: mockAdminUpdateUser,
    adminDeleteUser: mockAdminDeleteUser,
    adminResetUserPassword: mockAdminResetUserPassword,
  };
});

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useCreateUser", () => {
  describe("mutation call", () => {
    it("calls createUserAdmin with the provided body", async () => {
      mockCreateUserAdmin.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useCreateUser());

      const body = {
        body: {
          name: "Alice",
          username: "alice",
          email: "alice@example.com",
          password: "pass1",
        },
      };
      await act(() => result.current.mutateAsync(body as never));

      expect(mockCreateUserAdmin).toHaveBeenCalledWith(
        expect.objectContaining({ ...body, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      mockCreateUserAdmin.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useCreateUser());

      await act(() => result.current.mutateAsync({}));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("create failed");
      mockCreateUserAdmin.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useCreateUser());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      mockCreateUserAdmin.mockRejectedValue(new Error("create failed"));
      const { result } = renderHookWithClient(() => useCreateUser());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useUpdateUser", () => {
  describe("mutation call", () => {
    it("calls adminUpdateUser with path and body", async () => {
      mockAdminUpdateUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUpdateUser());

      const vars = { path: { id: "u1" }, body: { name: "Bob" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAdminUpdateUser).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful update", async () => {
      mockAdminUpdateUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUpdateUser());

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when update fails", async () => {
      const error = new Error("update failed");
      mockAdminUpdateUser.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useUpdateUser());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when update fails", async () => {
      mockAdminUpdateUser.mockRejectedValue(new Error("update failed"));
      const { result } = renderHookWithClient(() => useUpdateUser());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useDeleteUser", () => {
  describe("mutation call", () => {
    it("calls adminDeleteUser with the path", async () => {
      mockAdminDeleteUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useDeleteUser());

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAdminDeleteUser).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful delete", async () => {
      mockAdminDeleteUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useDeleteUser());

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when delete fails", async () => {
      const error = new Error("delete failed");
      mockAdminDeleteUser.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useDeleteUser());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when delete fails", async () => {
      mockAdminDeleteUser.mockRejectedValue(new Error("delete failed"));
      const { result } = renderHookWithClient(() => useDeleteUser());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useResetUserPassword", () => {
  describe("mutation call", () => {
    it("calls adminResetUserPassword with the path", async () => {
      mockAdminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "generated-pw" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars as never));

      expect(mockAdminResetUserPassword).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });

    it("returns the generated password from the mutation", async () => {
      mockAdminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "s3cr3t-pass" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      const data = await act(() => result.current.mutateAsync({} as never));

      expect(data).toEqual({ password: "s3cr3t-pass" });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful password reset", async () => {
      mockAdminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "pw" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      await act(() => result.current.mutateAsync({} as never));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when reset fails", async () => {
      const error = new Error("reset failed");
      mockAdminResetUserPassword.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useResetUserPassword());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when reset fails", async () => {
      mockAdminResetUserPassword.mockRejectedValue(new Error("reset failed"));
      const { result } = renderHookWithClient(() => useResetUserPassword());

      act(() => result.current.mutate({} as never));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
