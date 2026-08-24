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

const mockInvalidate = vi.fn();

const sdk = vi.hoisted(() =>
  mockSdkGen({
    createUserAdmin: vi.fn(),
    adminUpdateUser: vi.fn(),
    adminDeleteUser: vi.fn(),
    adminResetUserPassword: vi.fn(),
  }),
);

vi.mock("../useInvalidateQueries", () => ({
  useInvalidateByIds: vi.fn(() => mockInvalidate),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useCreateUser", () => {
  describe("mutation call", () => {
    it("calls createUserAdmin with the provided body", async () => {
      sdk.createUserAdmin.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useCreateUser());

      const body = {
        body: {
          name: "Alice",
          username: "alice",
          email: "alice@example.com",
          password: "pass1",
        },
      };
      await act(() => result.current.mutateAsync(body));

      expect(sdk.createUserAdmin).toHaveBeenCalledWith(
        expect.objectContaining({ ...body, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful mutation", async () => {
      sdk.createUserAdmin.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useCreateUser());

      await act(() => result.current.mutateAsync({}));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when mutation fails", async () => {
      const error = new Error("create failed");
      sdk.createUserAdmin.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useCreateUser());

      act(() => result.current.mutate({}));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when mutation fails", async () => {
      sdk.createUserAdmin.mockRejectedValue(new Error("create failed"));
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
      sdk.adminUpdateUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUpdateUser());

      const vars = { path: { id: "u1" }, body: { name: "Bob" } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.adminUpdateUser).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful update", async () => {
      sdk.adminUpdateUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useUpdateUser());

      await act(() => result.current.mutateAsync({ path: { id: "u1" } }));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when update fails", async () => {
      const error = new Error("update failed");
      sdk.adminUpdateUser.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useUpdateUser());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when update fails", async () => {
      sdk.adminUpdateUser.mockRejectedValue(new Error("update failed"));
      const { result } = renderHookWithClient(() => useUpdateUser());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useDeleteUser", () => {
  describe("mutation call", () => {
    it("calls adminDeleteUser with the path", async () => {
      sdk.adminDeleteUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useDeleteUser());

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.adminDeleteUser).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful delete", async () => {
      sdk.adminDeleteUser.mockResolvedValue(mockSdkResponse(undefined));
      const { result } = renderHookWithClient(() => useDeleteUser());

      await act(() => result.current.mutateAsync({ path: { id: "u1" } }));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when delete fails", async () => {
      const error = new Error("delete failed");
      sdk.adminDeleteUser.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useDeleteUser());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when delete fails", async () => {
      sdk.adminDeleteUser.mockRejectedValue(new Error("delete failed"));
      const { result } = renderHookWithClient(() => useDeleteUser());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});

describe("useResetUserPassword", () => {
  describe("mutation call", () => {
    it("calls adminResetUserPassword with the path", async () => {
      sdk.adminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "generated-pw" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      const vars = { path: { id: "u1" } };
      await act(() => result.current.mutateAsync(vars));

      expect(sdk.adminResetUserPassword).toHaveBeenCalledWith(
        expect.objectContaining({ ...vars, throwOnError: true }),
      );
    });

    it("returns the generated password from the mutation", async () => {
      sdk.adminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "s3cr3t-pass" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      const data = await act(() =>
        result.current.mutateAsync({ path: { id: "u1" } }),
      );

      expect(data).toEqual({ password: "s3cr3t-pass" });
    });
  });

  describe("on success", () => {
    it("calls invalidate after successful password reset", async () => {
      sdk.adminResetUserPassword.mockResolvedValue(
        mockSdkResponse({ password: "pw" }),
      );
      const { result } = renderHookWithClient(() => useResetUserPassword());

      await act(() => result.current.mutateAsync({ path: { id: "u1" } }));

      await waitFor(() => expect(mockInvalidate).toHaveBeenCalledTimes(1));
    });
  });

  describe("on failure", () => {
    it("exposes error when reset fails", async () => {
      const error = new Error("reset failed");
      sdk.adminResetUserPassword.mockRejectedValue(error);
      const { result } = renderHookWithClient(() => useResetUserPassword());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(result.current.error).toBe(error);
    });

    it("does not call invalidate when reset fails", async () => {
      sdk.adminResetUserPassword.mockRejectedValue(new Error("reset failed"));
      const { result } = renderHookWithClient(() => useResetUserPassword());

      act(() => result.current.mutate({ path: { id: "u1" } }));

      await waitFor(() => expect(result.current.isError).toBe(true));
      expect(mockInvalidate).not.toHaveBeenCalled();
    });
  });
});
