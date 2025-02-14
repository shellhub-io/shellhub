import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("Firewalls", () => {
  const firewalls = [
    {
      id: "5f1996c84d2190a22d5857bb",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      priority: 4,
      action: "allow",
      active: true,
      source_ip: "127.0.0.1",
      username: "shellhub",
      hostname: "shellhub",
    },
    {
      id: "5f1996c84d2190a22d5857cc",
      tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      priority: 3,
      action: "allow",
      active: false,
      source_ip: "127.0.0.1",
      username: "shellhub",
      hostname: "shellhub",
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
    hostname: "shellhub",
  };

  const numberFirewalls = 2;

  it("Return firewall default variables", () => {
    expect(store.getters["firewallRules/list"]).toEqual([]);
    expect(store.getters["firewallRules/get"]).toEqual({});
  });

  it("Verify initial state change for setFirewalls mutation", () => {
    store.commit("firewallRules/setFirewalls", { data: firewalls, headers: { "x-total-count": numberFirewalls } });
    expect(store.getters["firewallRules/list"]).toEqual(firewalls);
  });
  it("Verify inital state change for setSession mutation", () => {
    store.commit("firewallRules/setFirewall", { data: firewallRule });
    expect(store.getters["firewallRules/get"]).toEqual(firewallRule);
  });
});
