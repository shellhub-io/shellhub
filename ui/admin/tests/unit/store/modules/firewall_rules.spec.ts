import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import { IAdminFirewallRule } from "@admin/interfaces/IFirewallRule";
import { buildUrl } from "@tests/utils/url";

const mockFirewallRuleBase: IAdminFirewallRule = {
  id: "5f1996c84d2190a22d5857bb",
  tenant_id: "tenant-id-123",
  priority: 4,
  action: "allow",
  active: true,
  source_ip: "192.168.1.100",
  username: "admin",
  filter: { hostname: "admin-server", tags: [] },
};

describe("Admin Firewall Rules Store", () => {
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    firewallRulesStore = useFirewallRulesStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty firewall rules array", () => {
      expect(firewallRulesStore.firewallRules).toEqual([]);
    });

    it("should have zero firewall rules count", () => {
      expect(firewallRulesStore.firewallRulesCount).toBe(0);
    });
  });

  describe("fetchFirewallRulesList", () => {
    const baseUrl = "http://localhost:3000/admin/api/firewall/rules";

    it("should fetch firewall rules list successfully with default pagination", async () => {
      const rulesList = [mockFirewallRuleBase];

      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(200, rulesList, { "x-total-count": "1" });

      await expect(firewallRulesStore.fetchFirewallRulesList()).resolves.not.toThrow();

      expect(firewallRulesStore.firewallRules).toEqual(rulesList);
      expect(firewallRulesStore.firewallRulesCount).toBe(1);
    });

    it("should fetch firewall rules list successfully with custom pagination", async () => {
      const rulesList = [
        mockFirewallRuleBase,
        { ...mockFirewallRuleBase, id: "5f1996c84d2190a22d5857cc", priority: 3 },
      ];

      mockAdminApi.onGet(buildUrl(baseUrl, { page: "2", per_page: "20" })).reply(200, rulesList, { "x-total-count": "2" });

      await expect(firewallRulesStore.fetchFirewallRulesList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(firewallRulesStore.firewallRules).toEqual(rulesList);
      expect(firewallRulesStore.firewallRulesCount).toBe(2);
    });

    it("should fetch empty firewall rules list successfully", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(200, [], { "x-total-count": "0" });

      await expect(firewallRulesStore.fetchFirewallRulesList()).resolves.not.toThrow();

      expect(firewallRulesStore.firewallRules).toEqual([]);
      expect(firewallRulesStore.firewallRulesCount).toBe(0);
    });

    it("should throw on server error when fetching firewall rules list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).reply(500);

      await expect(firewallRulesStore.fetchFirewallRulesList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching firewall rules list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { page: "1", per_page: "10" })).networkError();

      await expect(firewallRulesStore.fetchFirewallRulesList()).rejects.toThrow("Network Error");
    });
  });

  describe("fetchFirewallRuleById", () => {
    const generateGetRuleUrl = (ruleId: string) => `http://localhost:3000/admin/api/firewall/rules/${ruleId}`;

    it("should fetch firewall rule by id successfully and return data", async () => {
      const ruleId = "5f1996c84d2190a22d5857bb";

      mockAdminApi.onGet(generateGetRuleUrl(ruleId)).reply(200, mockFirewallRuleBase);

      const result = await firewallRulesStore.fetchFirewallRuleById(ruleId);

      expect(result).toEqual(mockFirewallRuleBase);
    });

    it("should throw on not found error when fetching firewall rule by id", async () => {
      const ruleId = "non-existent-rule";

      mockAdminApi.onGet(generateGetRuleUrl(ruleId)).reply(404, { message: "Firewall rule not found" });

      await expect(firewallRulesStore.fetchFirewallRuleById(ruleId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching firewall rule by id", async () => {
      const ruleId = "5f1996c84d2190a22d5857bb";

      mockAdminApi.onGet(generateGetRuleUrl(ruleId)).networkError();

      await expect(firewallRulesStore.fetchFirewallRuleById(ruleId)).rejects.toThrow("Network Error");
    });
  });
});
