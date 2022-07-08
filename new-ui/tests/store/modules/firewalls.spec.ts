import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Firewalls', () => {
  const numberFirewalls = 2;

  const firewalls = [
    {
      id: '5f1996c84d2190a22d5857bb',
      tenant_id: '00000000-0000-4000-0000-000000000000',
      priority: 4,
      action: 'allow',
      active: true,
      source_ip: '127.0.0.1',
      username: 'shellhub',
      hostname: 'shellhub',
    },
    {
      id: '5f1996c84d2190a22d5857cc',
      tenant_id: '00000000-0000-4000-0000-000000000000',
      priority: 3,
      action: 'allow',
      active: false,
      source_ip: '127.0.0.1',
      username: 'shellhub',
      hostname: 'shellhub',
    },
  ];

  const firewallRule = {
    id: '5f1996c84d2190a22d5857bb',
    tenant_id: '00000000-0000-4000-0000-000000000000',
    priority: 4,
    action: 'allow',
    active: true,
    source_ip: '127.0.0.1',
    username: 'shellhub',
    hostname: 'shellhub',
  };

  const pagePerpageInitialValue = {
    page: 1,
    perPage: 10,
  };

  const pagePerpageValue = {
    page: 2,
    perPage: 50,
  };

  it('Return firewall default variables', () => {
    expect(store.getters['firewallRules/list']).toEqual([]);
    expect(store.getters['firewallRules/get']).toEqual({});
    expect(store.getters['firewallRules/getNumberFirewalls']).toEqual(0);
    expect(store.getters['firewallRules/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['firewallRules/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });

  it('Verify initial state change for setFirewalls mutation', () => {
    store.commit('firewallRules/setFirewalls', { data: firewalls, headers: { 'x-total-count': numberFirewalls } });
    expect(store.getters['firewallRules/list']).toEqual(firewalls);
    expect(store.getters['firewallRules/getNumberFirewalls']).toEqual(numberFirewalls);
  });
  it('Verify inital state change for setSession mutation', () => {
    store.commit('firewallRules/setFirewall', { data: firewallRule });
    expect(store.getters['firewallRules/get']).toEqual(firewallRule);
  });
  it('Verify inital state change for setPagePerpageFilter mutation', () => {
    store.commit('firewallRules/setPagePerpageFilter', pagePerpageValue);
    expect(store.getters['firewallRules/getPage']).toEqual(pagePerpageValue.page);
    expect(store.getters['firewallRules/getPerPage']).toEqual(pagePerpageValue.perPage);
  });
  it('Verify inital state change for resetPagePerpage mutation', () => {
    store.commit('firewallRules/resetPagePerpage');
    expect(store.getters['firewallRules/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['firewallRules/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });
  it('Verify remove firewall item from list for removeFirewalls mutation', () => {
    store.commit('firewallRules/removeFirewalls', firewallRule.id);
    expect(store.getters['firewallRules/list'].length).toEqual(numberFirewalls - 1);
  });
  it('Verify changed firewall object state for clearObjectFirewall mutation', () => {
    store.commit('firewallRules/clearObjectFirewalls');
    expect(store.getters['firewallRules/get']).toEqual({});
  });
  it('Verify changed firewall list state for clearListFirewalls mutation', () => {
    store.commit('firewallRules/clearListFirewalls');
    expect(store.getters['firewallRules/list']).toEqual([]);
    expect(store.getters['firewallRules/getNumberFirewalls']).toEqual(0);
  });
});
