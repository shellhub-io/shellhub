import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import {
  useNamespaceInvitations,
  useResolveInvitation,
} from "../useInvitations";
import type { MembershipInvitation } from "@/client";

const mockResolveFn = vi.fn();

vi.mock("@/client", () => ({
  getMembershipInvitationList: vi.fn(),
  getNamespaceMembershipInvitationList: vi.fn(),
  getMembershipInvitationListQueryKey: vi.fn((opts: unknown) => [
    { _id: "getMembershipInvitationList" },
    opts,
  ]),
  getNamespaceMembershipInvitationListQueryKey: vi.fn((opts: unknown) => [
    { _id: "getNamespaceMembershipInvitationList" },
    opts,
  ]),
  resolveInvitationOptions: vi.fn((opts: unknown) => ({
    queryKey: [{ _id: "resolveInvitation" }, opts],
    queryFn: () => mockResolveFn() as unknown,
  })),
}));

vi.mock("@/api/pagination", () => ({
  paginatedQueryFn: vi.fn(
    (_sdkFn: unknown, opts: { query: Record<string, unknown> }) => {
      return () => mockFetchFn(opts) as unknown;
    },
  ),
}));

const mockFetchFn = vi.fn();

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

describe("useNamespaceInvitations", () => {
  describe("returns", () => {
    it("returns invitations for the given tenant", async () => {
      const inv = makeInvitation({ status: "pending" });
      mockFetchFn.mockResolvedValue({ data: [inv], totalCount: 1 });

      const { result } = renderHookWithClient(() =>
        useNamespaceInvitations({ tenantId: "t1" }),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.invitations).toHaveLength(1);
    });

    it("returns totalCount from the paginated result", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 7 });

      const { result } = renderHookWithClient(() =>
        useNamespaceInvitations({ tenantId: "t1" }),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(7);
    });

    it("defaults invitations to empty array while loading", () => {
      mockFetchFn.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() =>
        useNamespaceInvitations({ tenantId: "t1" }),
      );

      expect(result.current.invitations).toEqual([]);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("fetch failed");
      mockFetchFn.mockRejectedValue(err);

      const { result } = renderHookWithClient(() =>
        useNamespaceInvitations({ tenantId: "t1" }),
      );

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(err);
    });
  });

  describe("enabled flag", () => {
    it("does not fetch when enabled is false", () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHookWithClient(() =>
        useNamespaceInvitations({ tenantId: "t1", enabled: false }),
      );

      expect(mockFetchFn).not.toHaveBeenCalled();
    });

    it("does not fetch when tenantId is empty", () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHookWithClient(() => useNamespaceInvitations({ tenantId: "" }));

      expect(mockFetchFn).not.toHaveBeenCalled();
    });
  });

  describe("pagination", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockFetchFn.mockResolvedValue({ data: [], totalCount: 0 });

      renderHookWithClient(() => useNamespaceInvitations({ tenantId: "t1" }));

      await waitFor(() => expect(mockFetchFn).toHaveBeenCalled());
      const [opts] = mockFetchFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });
  });
});

describe("useResolveInvitation", () => {
  it("normalizes wire fields to camelCase", async () => {
    mockResolveFn.mockResolvedValue({
      tenant_id: "t1",
      user_id: "u1",
      email: "alice@example.com",
      status: "confirmed",
    });

    const { result } = renderHookWithClient(() => useResolveInvitation("CODE"));

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.resolved).toEqual({
      tenantId: "t1",
      userId: "u1",
      email: "alice@example.com",
      status: "confirmed",
    });
  });

  it.each(["tenant_id", "user_id", "status"])(
    "returns null when %s is missing",
    async (field) => {
      const data = {
        tenant_id: "t1",
        user_id: "u1",
        email: "a@b.com",
        status: "confirmed",
        [field]: null,
      };
      mockResolveFn.mockResolvedValue(data);

      const { result } = renderHookWithClient(() =>
        useResolveInvitation("CODE"),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.resolved).toBeNull();
    },
  );

  it("falls back to empty string when email is null", async () => {
    mockResolveFn.mockResolvedValue({
      tenant_id: "t1",
      user_id: "u1",
      email: null,
      status: "invited",
    });

    const { result } = renderHookWithClient(() => useResolveInvitation("CODE"));

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.resolved?.email).toBe("");
  });

  it("does not fetch when invite is empty", () => {
    const { result } = renderHookWithClient(() => useResolveInvitation(""));

    expect(result.current.resolved).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(mockResolveFn).not.toHaveBeenCalled();
  });
});
