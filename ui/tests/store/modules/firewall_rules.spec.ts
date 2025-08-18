import { describe, expect, it, beforeEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { rulesApi } from "@/api/http";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

const firewallRuleData: IFirewallRule = {
  id: "rule1",
  tenant_id: "tenant1",
  priority: 1,
  action: "allow",
  active: true,
  source_ip: "192.168.1.0/24",
  username: "user1",
  filter: {
    hostname: ".*",
  },
};

describe("Firewall Rules Pinia Store", () => {
  setActivePinia(createPinia());
  const mockRulesApi = new MockAdapter(rulesApi.getAxios());
  const firewallRulesStore = useFirewallRulesStore();

  beforeEach(() => {
    firewallRulesStore.firewallRules = [];
    firewallRulesStore.firewallRuleCount = 0;
  });

  it("should have initial state values", () => {
    expect(firewallRulesStore.firewallRules).toEqual([]);
    expect(firewallRulesStore.firewallRuleCount).toBe(0);
  });

  it("should create firewall rule successfully", async () => {
    mockRulesApi.onPost("http://localhost:3000/api/firewall/rules").reply(200);

    await expect(firewallRulesStore.createFirewallRule(firewallRuleData)).resolves.not.toThrow();
  });

  it("should fetch firewall rule list successfully", async () => {
    const firewallRulesData = [
      firewallRuleData,
      {
        id: "rule2",
        tenant_id: "tenant1",
        priority: 2,
        action: "deny",
        active: true,
        source_ip: "10.0.0.0/8",
        username: "user2",
        filter: { hostname: "test.*" },
      },
    ];

    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, firewallRulesData, {
      "x-total-count": "2",
    });

    await firewallRulesStore.fetchFirewallRuleList({ page: 1, perPage: 10 });
    expect(firewallRulesStore.firewallRules).toEqual(firewallRulesData);
    expect(firewallRulesStore.firewallRuleCount).toBe(2);
  });

  it("should handle empty firewall rule list", async () => {
    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, [], {
      "x-total-count": "0",
    });

    await firewallRulesStore.fetchFirewallRuleList({ page: 1, perPage: 10 });

    expect(firewallRulesStore.firewallRules).toEqual([]);
    expect(firewallRulesStore.firewallRuleCount).toBe(0);
  });

  it("should use default pagination when no parameters provided", async () => {
    const firewallRulesData = [firewallRuleData];

    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, firewallRulesData, {
      "x-total-count": "1",
    });

    await firewallRulesStore.fetchFirewallRuleList();

    expect(firewallRulesStore.firewallRules).toEqual(firewallRulesData);
    expect(firewallRulesStore.firewallRuleCount).toBe(1);
  });

  it("should update firewall rule", async () => {
    const updatedFirewallRule: IFirewallRule = {
      ...firewallRuleData,
      active: false,
      action: "deny",
      filter: {
        hostname: "updated.*",
      },
    };

    mockRulesApi.onPut("http://localhost:3000/api/firewall/rules/rule1").reply(200);

    await expect(firewallRulesStore.updateFirewallRule(updatedFirewallRule)).resolves.not.toThrow();
  });

  it("should remove firewall rule", async () => {
    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/rule1").reply(200);
    await expect(firewallRulesStore.removeFirewallRule("rule1")).resolves.not.toThrow();
  });

  it("should handle fetch firewall rule list error", async () => {
    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(500);

    await expect(firewallRulesStore.fetchFirewallRuleList({ page: 1, perPage: 10 })).rejects.toThrow();

    expect(firewallRulesStore.firewallRules).toEqual([]);
    expect(firewallRulesStore.firewallRuleCount).toBe(0);
  });

  it("should handle create firewall rule error", async () => {
    mockRulesApi.onPost("http://localhost:3000/api/firewall/rules").reply(400);
    await expect(firewallRulesStore.createFirewallRule(firewallRuleData)).rejects.toThrow();
  });

  it("should handle update firewall rule error", async () => {
    mockRulesApi.onPut("http://localhost:3000/api/firewall/rules/rule1").reply(404);
    await expect(firewallRulesStore.updateFirewallRule(firewallRuleData)).rejects.toThrow();
  });

  it("should handle remove firewall rule error", async () => {
    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/rule1").reply(404);
    await expect(firewallRulesStore.removeFirewallRule("rule1")).rejects.toThrow();
  });
});
