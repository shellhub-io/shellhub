import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAdminUsers, useAdminUser } from "../useAdminUsers";
import { useAuthStore } from "../../stores/authStore";

// Mock the SDK functions used by the generated options/queryFn helpers.
vi.mock("../../client", () => ({
  getUsers: vi.fn(),
  getUser: vi.fn(),
}));

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  getUsersQueryKey: vi.fn((opts: unknown) => [{ _id: "getUsers" }, opts]),
  getUserOptions: vi.fn((opts: unknown) => ({
    queryKey: [{ _id: "getUser" }, opts],
    queryFn: mockGetUserFn,
  })),
}));

vi.mock("../../api/pagination", () => ({
  paginatedQueryFn: vi.fn(
    (_sdkFn: unknown, opts: { query: Record<string, unknown> }) => {
      return () => mockGetUsersFn(opts) as unknown;
    },
  ),
}));

const mockGetUsersFn = vi.fn();
const mockGetUserFn = vi.fn();

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
  // Default: authenticated admin
  useAuthStore.setState({ isAdmin: true } as never);
});

describe("useAdminUsers", () => {
  describe("when user is admin", () => {
    it("returns users from the paginated query result", async () => {
      const users = [
        { id: "u1", username: "alice" },
        { id: "u2", username: "bob" },
      ];
      mockGetUsersFn.mockResolvedValue({ data: users, totalCount: 2 });

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.users).toEqual(users);
    });

    it("returns totalCount from the paginated query result", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 99 });

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(99);
    });

    it("defaults users to empty array while loading", () => {
      // Never resolves — stays in loading state
      mockGetUsersFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      expect(result.current.users).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetUsersFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetUsersFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetUsersFn.mockRejectedValue(networkError);

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });

    it("exposes refetch function", () => {
      mockGetUsersFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      expect(typeof result.current.refetch).toBe("function");
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query", async () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminUsers(), {
        wrapper: createWrapper(),
      });

      // Query is disabled — stays in non-loading state with empty data
      expect(result.current.isLoading).toBe(false);
      expect(result.current.users).toEqual([]);
      expect(mockGetUsersFn).not.toHaveBeenCalled();
    });
  });

  describe("search filter", () => {
    it("passes search parameter to the query options", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 0 });

      const { result } = renderHook(() => useAdminUsers({ search: "alice" }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      // Query ran — the mock was called
      expect(mockGetUsersFn).toHaveBeenCalled();
    });

    it("does not pass filter when search is empty", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useAdminUsers({ search: "" }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockGetUsersFn).toHaveBeenCalled());
      // paginatedQueryFn receives options without a filter key when search is empty
      const [opts] = mockGetUsersFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.filter).toBeUndefined();
    });

    it("includes a base64-encoded filter when search is non-empty", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useAdminUsers({ search: "alice" }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockGetUsersFn).toHaveBeenCalled());
      const [opts] = mockGetUsersFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(typeof opts.query.filter).toBe("string");
      // Verify it decodes to valid JSON containing the search term
      const decoded = JSON.parse(
        atob(opts.query.filter as string),
      ) as unknown[];
      expect(JSON.stringify(decoded)).toContain("alice");
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useAdminUsers(), { wrapper: createWrapper() });

      await waitFor(() => expect(mockGetUsersFn).toHaveBeenCalled());
      const [opts] = mockGetUsersFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetUsersFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useAdminUsers({ page: 3, perPage: 25 }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockGetUsersFn).toHaveBeenCalled());
      const [opts] = mockGetUsersFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(3);
      expect(opts.query.per_page).toBe(25);
    });
  });
});

describe("useAdminUser", () => {
  describe("when user is admin", () => {
    it("returns query data for the given user id", async () => {
      const user = { id: "u1", username: "alice" };
      mockGetUserFn.mockResolvedValue(user);

      const { result } = renderHook(() => useAdminUser("u1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data).toEqual(user);
    });

    it("is loading initially when id is provided", () => {
      mockGetUserFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useAdminUser("u1"), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("not found");
      mockGetUserFn.mockRejectedValue(err);

      const { result } = renderHook(() => useAdminUser("u1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isError).toBe(true));
    });
  });

  describe("when id is empty", () => {
    it("does not execute the query", () => {
      const { result } = renderHook(() => useAdminUser(""), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(mockGetUserFn).not.toHaveBeenCalled();
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query even when id is provided", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminUser("u1"), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(mockGetUserFn).not.toHaveBeenCalled();
    });
  });
});
