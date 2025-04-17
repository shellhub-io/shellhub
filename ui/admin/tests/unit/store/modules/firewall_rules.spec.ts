import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";

describe("FirewallRules Store (Pinia)", () => {
  let firewallStore: ReturnType<typeof useFirewallRulesStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    firewallStore = useFirewallRulesStore();
  });

  const firewalls = [
    {
      id: "5f1996c84d2190a22d5857bb",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      priority: 4,
      action: "allow",
      active: true,
      source_ip: "127.0.0.1",
      username: "shellhub",
      filter: { hostname: "shellhub", tags: [] },
    },
    {
      id: "5f1996c84d2190a22d5857cc",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      priority: 3,
      action: "deny",
      active: false,
      source_ip: "127.0.0.1",
      username: "shellhub",
      filter: { hostname: "shellhub", tags: [] },
    },
  ];

  const firewallRule = {
    id: "5f1996c84d2190a22d5857bb",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 4,
    action: "allow",
    active: true,
    source_ip: "127.0.0.1",
    username: "shellhub",
    filter: { hostname: "shellhub", tags: [] },
  };

  it("returns default values", () => {
    expect(firewallStore.list).toEqual([]);
    expect(firewallStore.getFirewall).toEqual({});
    expect(firewallStore.getNumberFirewalls).toEqual(0);
  });

  it("sets firewalls and total count", () => {
    firewallStore.firewalls = firewalls;
    firewallStore.numberFirewalls = firewalls.length;

    expect(firewallStore.list).toEqual(firewalls);
    expect(firewallStore.getNumberFirewalls).toBe(2);
  });

  it("sets a single firewall rule", () => {
    firewallStore.firewall = firewallRule;
    expect(firewallStore.getFirewall).toEqual(firewallRule);
  });
});
