import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { decodeB64url } from "@/tests/decodeB64url";
import { useAdminUsers, useAdminUser } from "../useAdminUsers";
import { useAuthStore } from "@/stores/authStore";

const mockGetUsers = vi.hoisted(() => vi.fn());
const mockGetUser = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getUsers: mockGetUsers, getUser: mockGetUser };
});

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true } as never);
});

describe("useAdminUsers", () => {
  describe("when user is admin", () => {
    it("returns users from the paginated query result", async () => {
      const users = [
        { id: "u1", username: "alice" },
        { id: "u2", username: "bob" },
      ];
      mockGetUsers.mockResolvedValue(
        mockSdkResponse(users, { "X-Total-Count": "2" }),
      );

      const { result } = renderHookWithClient(() => useAdminUsers());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.users).toEqual(users);
    });

    it("returns totalCount from the X-Total-Count header", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "99" }),
      );

      const { result } = renderHookWithClient(() => useAdminUsers());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(99);
    });

    it("defaults users to empty array while loading", () => {
      mockGetUsers.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminUsers());

      expect(result.current.users).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetUsers.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminUsers());

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetUsers.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminUsers());

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetUsers.mockRejectedValue(networkError);

      const { result } = renderHookWithClient(() => useAdminUsers());

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });

    it("exposes refetch function", () => {
      mockGetUsers.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminUsers());

      expect(typeof result.current.refetch).toBe("function");
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHookWithClient(() => useAdminUsers());

      expect(result.current.isLoading).toBe(false);
      expect(result.current.users).toEqual([]);
      expect(mockGetUsers).not.toHaveBeenCalled();
    });
  });

  describe("search filter", () => {
    it("passes search parameter to the query options", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      const { result } = renderHookWithClient(() =>
        useAdminUsers({ search: "alice" }),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(mockGetUsers).toHaveBeenCalled();
    });

    it("does not pass filter when search is empty", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => useAdminUsers({ search: "" }));

      await waitFor(() => expect(mockGetUsers).toHaveBeenCalled());
      const [opts] = mockGetUsers.mock.calls[0];
      expect(opts.query.filter).toBeUndefined();
    });

    it("includes a base64-encoded filter when search is non-empty", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => useAdminUsers({ search: " >" }));

      await waitFor(() => expect(mockGetUsers).toHaveBeenCalled());
      const [opts] = mockGetUsers.mock.calls[0];
      expect(typeof opts.query.filter).toBe("string");
      const decoded = decodeB64url(opts.query.filter as string) as unknown[];
      expect(JSON.stringify(decoded)).toContain(" >");
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => useAdminUsers());

      await waitFor(() => expect(mockGetUsers).toHaveBeenCalled());
      const [opts] = mockGetUsers.mock.calls[0];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetUsers.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => useAdminUsers({ page: 3, perPage: 25 }));

      await waitFor(() => expect(mockGetUsers).toHaveBeenCalled());
      const [opts] = mockGetUsers.mock.calls[0];
      expect(opts.query.page).toBe(3);
      expect(opts.query.per_page).toBe(25);
    });
  });
});

describe("useAdminUser", () => {
  describe("when user is admin", () => {
    it("returns query data for the given user id", async () => {
      const user = { id: "u1", username: "alice" };
      mockGetUser.mockResolvedValue(mockSdkResponse(user));

      const { result } = renderHookWithClient(() => useAdminUser("u1"));

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data).toEqual(user);
    });

    it("is loading initially when id is provided", () => {
      mockGetUser.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminUser("u1"));

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("not found");
      mockGetUser.mockRejectedValue(err);

      const { result } = renderHookWithClient(() => useAdminUser("u1"));

      await waitFor(() => expect(result.current.isError).toBe(true));
    });
  });

  describe("when id is empty", () => {
    it("does not execute the query", () => {
      const { result } = renderHookWithClient(() => useAdminUser(""));

      expect(result.current.isLoading).toBe(false);
      expect(mockGetUser).not.toHaveBeenCalled();
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query even when id is provided", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHookWithClient(() => useAdminUser("u1"));

      expect(result.current.isLoading).toBe(false);
      expect(mockGetUser).not.toHaveBeenCalled();
    });
  });
});
