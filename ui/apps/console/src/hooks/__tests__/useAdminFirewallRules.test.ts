import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import {
  useAdminFirewallRules,
  useAdminFirewallRule,
} from "../useAdminFirewallRules";
import { useAuthStore } from "@/stores/authStore";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";

const mockGetFirewallRulesAdmin = vi.hoisted(() => vi.fn());
const mockGetFirewallRuleAdmin = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    getFirewallRulesAdmin: mockGetFirewallRulesAdmin,
    getFirewallRuleAdmin: mockGetFirewallRuleAdmin,
  };
});

beforeEach(() => {
  vi.clearAllMocks();
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
          filter: { hostname: ".*", tags: [] },
        },
        {
          id: "rule-2",
          tenant_id: "tenant-abc",
          priority: 2,
          action: "deny",
          active: false,
          source_ip: "192.168.1.0/24",
          username: "admin",
          filter: { hostname: "my-host", tags: [] },
        },
      ];
      mockGetFirewallRulesAdmin.mockResolvedValue(
        mockSdkResponse(rules, { "X-Total-Count": "2" }),
      );

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.rules).toHaveLength(2);
      expect(result.current.rules[0].id).toBe("rule-1");
      expect(result.current.rules[1].id).toBe("rule-2");
    });

    it("returns totalCount from the X-Total-Count header", async () => {
      mockGetFirewallRulesAdmin.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "42" }),
      );

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.totalCount).toBe(42);
    });

    it("defaults rules to empty array while loading", () => {
      mockGetFirewallRulesAdmin.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      expect(result.current.rules).toEqual([]);
    });

    it("defaults totalCount to 0 while loading", () => {
      mockGetFirewallRulesAdmin.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      expect(result.current.totalCount).toBe(0);
    });

    it("returns isLoading true initially", () => {
      mockGetFirewallRulesAdmin.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const networkError = new Error("network failure");
      mockGetFirewallRulesAdmin.mockRejectedValue(networkError);

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      await waitFor(() => expect(result.current.error).toBeTruthy());
      expect(result.current.error).toBe(networkError);
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHookWithClient(() => useAdminFirewallRules());

      expect(result.current.isLoading).toBe(false);
      expect(result.current.rules).toEqual([]);
      expect(mockGetFirewallRulesAdmin).not.toHaveBeenCalled();
    });
  });

  describe("pagination defaults", () => {
    it("uses page 1 and perPage 10 as defaults", async () => {
      mockGetFirewallRulesAdmin.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() => useAdminFirewallRules());

      await waitFor(() => expect(mockGetFirewallRulesAdmin).toHaveBeenCalled());
      const [opts] = mockGetFirewallRulesAdmin.mock.calls[0];
      expect(opts.query.page).toBe(1);
      expect(opts.query.per_page).toBe(10);
    });

    it("forwards custom page and perPage", async () => {
      mockGetFirewallRulesAdmin.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      renderHookWithClient(() =>
        useAdminFirewallRules({ page: 3, perPage: 25 }),
      );

      await waitFor(() => expect(mockGetFirewallRulesAdmin).toHaveBeenCalled());
      const [opts] = mockGetFirewallRulesAdmin.mock.calls[0];
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
        filter: { hostname: ".*", tags: [] },
      };
      mockGetFirewallRuleAdmin.mockResolvedValue(mockSdkResponse(rawRule));

      const { result } = renderHookWithClient(() =>
        useAdminFirewallRule("rule-1"),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data?.id).toBe("rule-1");
    });

    it("passes through the filter unchanged", async () => {
      const rawRule = {
        id: "rule-1",
        tenant_id: "tenant-abc",
        priority: 1,
        action: "allow",
        active: true,
        source_ip: ".*",
        username: ".*",
        filter: { hostname: "my-host", tags: [] },
      };
      mockGetFirewallRuleAdmin.mockResolvedValue(mockSdkResponse(rawRule));

      const { result } = renderHookWithClient(() =>
        useAdminFirewallRule("rule-1"),
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.data?.filter).toEqual({
        hostname: "my-host",
        tags: [],
      });
    });

    it("is loading initially when id is provided", () => {
      mockGetFirewallRuleAdmin.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() =>
        useAdminFirewallRule("rule-1"),
      );

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes error when query fails", async () => {
      const err = new Error("not found");
      mockGetFirewallRuleAdmin.mockRejectedValue(err);

      const { result } = renderHookWithClient(() =>
        useAdminFirewallRule("rule-1"),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));
    });
  });

  describe("when id is empty", () => {
    it("does not execute the query", () => {
      const { result } = renderHookWithClient(() => useAdminFirewallRule(""));

      expect(result.current.isLoading).toBe(false);
      expect(mockGetFirewallRuleAdmin).not.toHaveBeenCalled();
    });
  });

  describe("when user is not admin", () => {
    it("does not execute the query even when id is provided", () => {
      useAuthStore.setState({ isAdmin: false } as never);

      const { result } = renderHookWithClient(() =>
        useAdminFirewallRule("rule-1"),
      );

      expect(result.current.isLoading).toBe(false);
      expect(mockGetFirewallRuleAdmin).not.toHaveBeenCalled();
    });
  });
});
