import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useUserInvitations, useNamespaceInvitations } from "../useInvitations";
import type { MembershipInvitation } from "../../client";

vi.mock("../../client", () => ({
  getMembershipInvitationList: vi.fn(),
  getNamespaceMembershipInvitationList: vi.fn(),
}));

vi.mock("../../client/@tanstack/react-query.gen", () => ({
  getMembershipInvitationListQueryKey: vi.fn(
    (opts: unknown) => [{ _id: "getMembershipInvitationList" }, opts],
  ),
  getNamespaceMembershipInvitationListQueryKey: vi.fn(
    (opts: unknown) => [{ _id: "getNamespaceMembershipInvitationList" }, opts],
  ),
}));

vi.mock("../../api/pagination", () => ({
  paginatedQueryFn: vi.fn(
    (_sdkFn: unknown, opts: { query: Record<string, unknown> }) => {
      return () => mockFetchFn(opts) as unknown;
    },
  ),
}));

const mockFetchFn = vi.fn();

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function makeInvitation(
  overrides: Partial<MembershipInvitation> = {},
): MembershipInvitation {
  return {
    namespace: { tenant_id: "t1", name: "my-ns" },
    user: { id: "u1", email: "alice@example.com" },
    invited_by: "owner@example.com",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    expires_at: "2024-01-08T00:00:00Z",
    status: "pending",
    status_updated_at: "2024-01-01T00:00:00Z",
    role: "operator",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useUserInvitations", () => {
  describe("returns", () => {
    it("returns invitations from the paginated result", async () => {
      const inv = makeInvitation({ status: "pending" });
      mockFetchFn.mockResolvedValue({ data: [inv], totalCount: 1 });

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.invitations).toHaveLength(1);
      expect(result.current.invitations[0].user.email).toBe("alice@example.com");
    });

    it("returns totalCount from the paginated result", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 42 });

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(42);
    });

    it("defaults invitations to empty array while loading", () => {
      mockFetchFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      expect(result.current.invitations).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockFetchFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockFetchFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockFetchFn.mockRejectedValue(networkError);

      const { result } = renderHook(() => useUserInvitations(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useUserInvitations(), { wrapper: createWrapper() });

      await waitFor(() => expect(mockFetchFn).toHaveBeenCalled());
      const [opts] = mockFetchFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useUserInvitations({ page: 2, perPage: 25 }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockFetchFn).toHaveBeenCalled());
      const [opts] = mockFetchFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(2);
      expect(opts.query.per_page).toBe(25);
    });
  });

  describe("enabled flag", () => {
    it("does not fetch when enabled is false", () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useUserInvitations({ enabled: false }), {
        wrapper: createWrapper(),
      });

      expect(mockFetchFn).not.toHaveBeenCalled();
    });
  });
});

describe("useNamespaceInvitations", () => {
  describe("returns", () => {
    it("returns invitations for the given tenant", async () => {
      const inv = makeInvitation({ status: "pending" });
      mockFetchFn.mockResolvedValue({ data: [inv], totalCount: 1 });

      const { result } = renderHook(
        () => useNamespaceInvitations({ tenantId: "t1" }),
        { wrapper: createWrapper() },
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.invitations).toHaveLength(1);
    });

    it("returns totalCount from the paginated result", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 7 });

      const { result } = renderHook(
        () => useNamespaceInvitations({ tenantId: "t1" }),
        { wrapper: createWrapper() },
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(7);
    });

    it("defaults invitations to empty array while loading", () => {
      mockFetchFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(
        () => useNamespaceInvitations({ tenantId: "t1" }),
        { wrapper: createWrapper() },
      );

      expect(result.current.invitations).toEqual([]);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("fetch failed");
      mockFetchFn.mockRejectedValue(err);

      const { result } = renderHook(
        () => useNamespaceInvitations({ tenantId: "t1" }),
        { wrapper: createWrapper() },
      );

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(err);
    });
  });

  describe("enabled flag", () => {
    it("does not fetch when enabled is false", () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useNamespaceInvitations({ tenantId: "t1", enabled: false }), {
        wrapper: createWrapper(),
      });

      expect(mockFetchFn).not.toHaveBeenCalled();
    });

    it("does not fetch when tenantId is empty", () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useNamespaceInvitations({ tenantId: "" }), {
        wrapper: createWrapper(),
      });

      expect(mockFetchFn).not.toHaveBeenCalled();
    });
  });

  describe("pagination", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHook(() => useNamespaceInvitations({ tenantId: "t1" }), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(mockFetchFn).toHaveBeenCalled());
      const [opts] = mockFetchFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });
  });
});
