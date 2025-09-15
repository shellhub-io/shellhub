import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";

const mockFirewallRules = [
  {
    id: "5f1996c84d2190a22d5857bb",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 4,
    action: "allow" as const,
    active: true,
    source_ip: "127.0.0.1",
    username: "shellhub",
    filter: { hostname: "shellhub", tags: [] },
  },
  {
    id: "5f1996c84d2190a22d5857cc",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 3,
    action: "deny" as const,
    active: false,
    source_ip: "127.0.0.1",
    username: "shellhub",
    filter: { hostname: "shellhub", tags: [] },
  },
];

describe("Firewall Rules Store", () => {
  setActivePinia(createPinia());
  const firewallRulesStore = useFirewallRulesStore();

  it("returns default values", () => {
    expect(firewallRulesStore.firewallRules).toEqual([]);
    expect(firewallRulesStore.firewallRulesCount).toEqual(0);
  });

  it("sets firewalls and total count", () => {
    firewallRulesStore.firewallRules = mockFirewallRules;
    firewallRulesStore.firewallRulesCount = mockFirewallRules.length;

    expect(firewallRulesStore.firewallRules).toEqual(mockFirewallRules);
    expect(firewallRulesStore.firewallRulesCount).toBe(2);
  });
});
