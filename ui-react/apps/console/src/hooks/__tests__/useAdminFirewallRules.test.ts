import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminFirewallRules,
  useAdminFirewallRule,
} from "../useAdminFirewallRules";
import { useAuthStore } from "@/stores/authStore";

// Mock the SDK functions used by the generated options/queryFn helpers.
vi.mock("@/client", () => ({
  getFirewallRulesAdmin: vi.fn(),
  getFirewallRuleAdmin: vi.fn(),
}));

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getFirewallRulesAdminQueryKey: vi.fn((opts: unknown) => [
    { _id: "getFirewallRulesAdmin" },
    opts,
  ]),
  getFirewallRuleAdminOptions: vi.fn((opts: unknown) => ({
    queryKey: [{ _id: "getFirewallRuleAdmin" }, opts],
    queryFn: mockGetFirewallRuleAdminFn,
  })),
}));

vi.mock("@/api/pagination", () => ({
  paginatedQueryFn: vi.fn(
    (_sdkFn: unknown, opts: { query: Record<string, unknown> }) => {
      return () => mockGetFirewallRulesAdminFn(opts) as unknown;
    },
  ),
}));

const mockGetFirewallRulesAdminFn = vi.fn();
const mockGetFirewallRuleAdminFn = vi.fn();

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

describe("useAdminFirewallRules", () => {
  describe("when user is admin", () => {
    it("returns rules from the paginated query result", async () => {
      const rules = [
        {
          id: "rule-1",
          tenant_id: "tenant-abc",
          priority: 1,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: ".*" },
        },
        {
          id: "rule-2",
          tenant_id: "tenant-abc",
          priority: 2,
          action: "deny",
          active: false,
          source_ip: "192.168.1.0/24",
          username: "admin",
          filter: { hostname: "my-host" },
        },
      ];
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: rules,
        totalCount: 2,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules).toHaveLength(2);
      expect(result.current.rules[0].id).toBe("rule-1");
      expect(result.current.rules[1].id).toBe("rule-2");
    });

    it("normalizes hostname filter from raw rule", async () => {
      const rules = [
        {
          id: "rule-1",
          tenant_id: "tenant-abc",
          priority: 1,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: "my-host" },
        },
      ];
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: rules,
        totalCount: 1,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules[0].filter).toEqual({ hostname: "my-host" });
    });

    it("normalizes tags filter from raw rule", async () => {
      const rules = [
        {
          id: "rule-1",
          tenant_id: "tenant-abc",
          priority: 1,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { tags: [{ name: "production" }, "staging"] },
        },
      ];
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: rules,
        totalCount: 1,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules[0].filter).toEqual({
        tags: ["production", "staging"],
      });
    });

    it("treats empty tags array as wildcard hostname filter", async () => {
      const rules = [
        {
          id: "rule-1",
          tenant_id: "t",
          priority: 1,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { tags: [] },
        },
      ];
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: rules,
        totalCount: 1,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules[0].filter).toEqual({ hostname: ".*" });
    });

    it("returns totalCount from the paginated query result", async () => {
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: [],
        totalCount: 42,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(42);
    });

    it("defaults rules to empty array while loading", () => {
      // Never resolves — stays in loading state
      mockGetFirewallRulesAdminFn.mockReturnValue(new Promise(() => { }));

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      expect(result.current.rules).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetFirewallRulesAdminFn.mockReturnValue(new Promise(() => { }));

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetFirewallRulesAdminFn.mockReturnValue(new Promise(() => { }));

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetFirewallRulesAdminFn.mockRejectedValue(networkError);

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });

    it("filters out rules without an id", async () => {
      const rules = [
        {
          id: "rule-1",
          tenant_id: "tenant-abc",
          priority: 1,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: ".*" },
        },
        // Rule without an id — should be filtered out
        {
          tenant_id: "tenant-abc",
          priority: 2,
          action: "deny",
          active: false,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: ".*" },
        },
      ];
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: rules,
        totalCount: 2,
      });

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules).toHaveLength(1);
      expect(result.current.rules[0].id).toBe("rule-1");
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query", async () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminFirewallRules(), {
        wrapper: createWrapper(),
      });

      // Query is disabled — stays in non-loading state with empty data
      expect(result.current.isLoading).toBe(false);
      expect(result.current.rules).toEqual([]);
      expect(mockGetFirewallRulesAdminFn).not.toHaveBeenCalled();
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: [],
        totalCount: 0,
      });

      renderHook(() => useAdminFirewallRules(), { wrapper: createWrapper() });

      await waitFor(() =>
        expect(mockGetFirewallRulesAdminFn).toHaveBeenCalled(),
      );
      const [opts] = mockGetFirewallRulesAdminFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetFirewallRulesAdminFn.mockResolvedValue({
        data: [],
        totalCount: 0,
      });

      renderHook(() => useAdminFirewallRules({ page: 3, perPage: 25 }), {
        wrapper: createWrapper(),
      });

      await waitFor(() =>
        expect(mockGetFirewallRulesAdminFn).toHaveBeenCalled(),
      );
      const [opts] = mockGetFirewallRulesAdminFn.mock.calls[0] as [
        { query: Record<string, unknown> },
      ];
      expect(opts.query.page).toBe(3);
      expect(opts.query.per_page).toBe(25);
    });
  });
});

describe("useAdminFirewallRule", () => {
  describe("when user is admin", () => {
    it("returns query data for the given rule id", async () => {
      const rawRule = {
        id: "rule-1",
        tenant_id: "tenant-abc",
        priority: 1,
        action: "allow",
        active: true,
        source_ip: ".*",
        username: ".*",
        filter: { hostname: ".*" },
      };
      mockGetFirewallRuleAdminFn.mockResolvedValue(rawRule);

      const { result } = renderHook(() => useAdminFirewallRule("rule-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data?.id).toBe("rule-1");
    });

    it("applies normalization via select — returns FirewallRule shape", async () => {
      const rawRule = {
        id: "rule-1",
        tenant_id: "tenant-abc",
        priority: 1,
        action: "allow",
        active: true,
        source_ip: ".*",
        username: ".*",
        filter: { hostname: "my-host" },
      };
      mockGetFirewallRuleAdminFn.mockResolvedValue(rawRule);

      const { result } = renderHook(() => useAdminFirewallRule("rule-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data?.filter).toEqual({ hostname: "my-host" });
    });

    it("is loading initially when id is provided", () => {
      mockGetFirewallRuleAdminFn.mockReturnValue(new Promise(() => { }));

      const { result } = renderHook(() => useAdminFirewallRule("rule-1"), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("not found");
      mockGetFirewallRuleAdminFn.mockRejectedValue(err);

      const { result } = renderHook(() => useAdminFirewallRule("rule-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isError).toBe(true));
    });
  });

  describe("when id is empty", () => {
    it("does not execute the query", () => {
      const { result } = renderHook(() => useAdminFirewallRule(""), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(mockGetFirewallRuleAdminFn).not.toHaveBeenCalled();
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query even when id is provided", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHook(() => useAdminFirewallRule("rule-1"), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(mockGetFirewallRuleAdminFn).not.toHaveBeenCalled();
    });
  });
});
