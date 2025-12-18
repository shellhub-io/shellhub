import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { rulesApi } from "@/api/http";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { buildUrl } from "../../utils/url";

const mockFirewallRuleBase: IFirewallRule = {
  id: "rule-123",
  tenant_id: "tenant-456",
  priority: 1,
  action: "allow",
  active: true,
  source_ip: "192.168.1.0/24",
  username: "testuser",
  filter: {
    hostname: ".*",
  },
};

describe("Firewall Rules Store", () => {
  let mockRulesApi: MockAdapter;
  let store: ReturnType<typeof useFirewallRulesStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockRulesApi = new MockAdapter(rulesApi.getAxios());
    store = useFirewallRulesStore();
  });

  afterEach(() => { mockRulesApi.reset(); });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.firewallRules).toEqual([]);
      expect(store.firewallRuleCount).toBe(0);
    });
  });

  describe("fetchFirewallRuleList", () => {
    const baseUrl = "http://localhost:3000/api/firewall/rules";

    it("should fetch firewall rules successfully with pagination", async () => {
      const mockRules = [
        mockFirewallRuleBase,
        {
          ...mockFirewallRuleBase,
          id: "rule-456",
          priority: 2,
          action: "deny" as const,
          source_ip: "10.0.0.0/8",
          username: "admin",
        },
      ];

      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, mockRules, {
          "x-total-count": "2",
        });

      await store.fetchFirewallRuleList({ page: 1, perPage: 10 });

      expect(store.firewallRules).toEqual(mockRules);
      expect(store.firewallRuleCount).toBe(2);
    });

    it("should handle empty firewall rules list", async () => {
      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchFirewallRuleList({ page: 1, perPage: 10 });

      expect(store.firewallRules).toEqual([]);
      expect(store.firewallRuleCount).toBe(0);
    });

    it("should use default pagination when no parameters provided", async () => {
      const mockRules = [mockFirewallRuleBase];

      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, mockRules, {
          "x-total-count": "1",
        });

      await store.fetchFirewallRuleList();

      expect(store.firewallRules).toEqual(mockRules);
      expect(store.firewallRuleCount).toBe(1);
    });

    it("should fetch rules with different pagination values", async () => {
      const mockRules = [mockFirewallRuleBase];

      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "2", per_page: "25" }))
        .reply(200, mockRules, {
          "x-total-count": "1",
        });

      await store.fetchFirewallRuleList({ page: 2, perPage: 25 });

      expect(store.firewallRules).toEqual(mockRules);
      expect(store.firewallRuleCount).toBe(1);
    });

    it("should reset state when request fails with forbidden error", async () => {
      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.fetchFirewallRuleList({ page: 1, perPage: 10 }),
      ).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.firewallRules).toEqual([]);
      expect(store.firewallRuleCount).toBe(0);
    });

    it("should reset state when network error occurs", async () => {
      mockRulesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .networkError();

      await expect(
        store.fetchFirewallRuleList({ page: 1, perPage: 10 }),
      ).rejects.toThrow();

      expect(store.firewallRules).toEqual([]);
      expect(store.firewallRuleCount).toBe(0);
    });
  });

  describe("createFirewallRule", () => {
    const createRuleUrl = "http://localhost:3000/api/firewall/rules";

    it("should create firewall rule successfully", async () => {
      mockRulesApi
        .onPost(createRuleUrl)
        .reply(200);

      await expect(store.createFirewallRule(mockFirewallRuleBase)).resolves.not.toThrow();
    });

    it("should create deny rule successfully", async () => {
      const denyRule = {
        ...mockFirewallRuleBase,
        action: "deny" as const,
        priority: 2,
      };

      mockRulesApi
        .onPost(createRuleUrl)
        .reply(200);

      await expect(store.createFirewallRule(denyRule)).resolves.not.toThrow();
    });

    it("should handle validation error when creating rule", async () => {
      mockRulesApi
        .onPost(createRuleUrl)
        .reply(400, { message: "Invalid rule data" });

      await expect(
        store.createFirewallRule(mockFirewallRuleBase),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when creating rule", async () => {
      mockRulesApi
        .onPost(createRuleUrl)
        .networkError();

      await expect(store.createFirewallRule(mockFirewallRuleBase)).rejects.toThrow();
    });
  });

  describe("updateFirewallRule", () => {
    it("should update firewall rule successfully", async () => {
      const updatedRule = {
        ...mockFirewallRuleBase,
        active: false,
        priority: 5,
      };

      mockRulesApi
        .onPut(`http://localhost:3000/api/firewall/rules/${updatedRule.id}`)
        .reply(200);

      await expect(store.updateFirewallRule(updatedRule)).resolves.not.toThrow();
    });

    it("should update rule action successfully", async () => {
      const updatedRule = {
        ...mockFirewallRuleBase,
        action: "deny" as const,
      };

      mockRulesApi
        .onPut(`http://localhost:3000/api/firewall/rules/${updatedRule.id}`)
        .reply(200);

      await expect(store.updateFirewallRule(updatedRule)).resolves.not.toThrow();
    });

    it("should update rule filter successfully", async () => {
      const updatedRule = {
        ...mockFirewallRuleBase,
        filter: {
          hostname: "prod-.*",
        },
      };

      mockRulesApi
        .onPut(`http://localhost:3000/api/firewall/rules/${updatedRule.id}`)
        .reply(200);

      await expect(store.updateFirewallRule(updatedRule)).resolves.not.toThrow();
    });

    it("should handle not found error when updating rule", async () => {
      mockRulesApi
        .onPut(`http://localhost:3000/api/firewall/rules/${mockFirewallRuleBase.id}`)
        .reply(404, { message: "Firewall rule not found" });

      await expect(
        store.updateFirewallRule(mockFirewallRuleBase),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when updating rule", async () => {
      mockRulesApi
        .onPut(`http://localhost:3000/api/firewall/rules/${mockFirewallRuleBase.id}`)
        .networkError();

      await expect(store.updateFirewallRule(mockFirewallRuleBase)).rejects.toThrow();
    });
  });

  describe("removeFirewallRule", () => {
    it("should remove firewall rule successfully", async () => {
      mockRulesApi
        .onDelete("http://localhost:3000/api/firewall/rules/rule-123")
        .reply(200);

      await expect(store.removeFirewallRule("rule-123")).resolves.not.toThrow();
    });

    it("should handle not found error when removing rule", async () => {
      mockRulesApi
        .onDelete("http://localhost:3000/api/firewall/rules/rule-123")
        .reply(404, { message: "Firewall rule not found" });

      await expect(
        store.removeFirewallRule("rule-123"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when removing rule", async () => {
      mockRulesApi
        .onDelete("http://localhost:3000/api/firewall/rules/rule-123")
        .networkError();

      await expect(store.removeFirewallRule("rule-123")).rejects.toThrow();
    });
  });
});
