import store from '@/store';

describe('Firewalls', () => {
  const numberFirewalls = 2;
  const firewalls = [
    {
      id: '5f1996c84d2190a22d5857bb',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      priority: 4,
      action: 'allow',
      active: true,
      source_ip: '127.0.0.1',
      username: 'shellhub',
      hostname: 'shellhub',
    },
    {
      id: '5f1996c84d2190a22d5857cc',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
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
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    priority: 4,
    action: 'allow',
    active: true,
    source_ip: '127.0.0.1',
    username: 'shellhub',
    hostname: 'shellhub',
  };
  const pagePerpageInitialValue = {
    page: 0,
    perPage: 10,
  };
  const pagePerpageValue = {
    page: 1,
    perPage: 50,
  };

  it('Return firewall default variables', () => {
    expect(store.getters['firewallrules/list']).toEqual([]);
    expect(store.getters['firewallrules/get']).toEqual({});
    expect(store.getters['firewallrules/getNumberFirewalls']).toEqual(0);
    expect(store.getters['firewallrules/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['firewallrules/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });

  it('Verify initial state change for setFirewalls mutation', () => {
    store.commit('firewallrules/setFirewalls', { data: firewalls, headers: { 'x-total-count': numberFirewalls } });
    expect(store.getters['firewallrules/list']).toEqual(firewalls);
    expect(store.getters['firewallrules/getNumberFirewalls']).toEqual(numberFirewalls);
  });
  it('Verify inital state change for setSession mutation', () => {
    store.commit('firewallrules/setFirewall', { data: firewallRule });
    expect(store.getters['firewallrules/get']).toEqual(firewallRule);
  });
  it('Verify inital state change for setPagePerpageFilter mutation', () => {
    store.commit('firewallrules/setPagePerpageFilter', pagePerpageValue);
    expect(store.getters['firewallrules/getPage']).toEqual(pagePerpageValue.page);
    expect(store.getters['firewallrules/getPerPage']).toEqual(pagePerpageValue.perPage);
  });
  it('Verify inital state change for resetPagePerpage mutation', () => {
    store.commit('firewallrules/resetPagePerpage');
    expect(store.getters['firewallrules/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['firewallrules/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });
  it('Verify remove firewall item from list for removeFirewalls mutation', () => {
    store.commit('firewallrules/removeFirewalls', firewallRule.id);
    expect(store.getters['firewallrules/list'].length).toEqual(numberFirewalls - 1);
  });
  it('Verify changed firewall object state for clearObjectFirewall mutation', () => {
    store.commit('firewallrules/clearObjectFirewalls');
    expect(store.getters['firewallrules/get']).toEqual({});
  });
  it('Verify changed firewall list state for clearListFirewalls mutation', () => {
    store.commit('firewallrules/clearListFirewalls');
    expect(store.getters['firewallrules/list']).toEqual([]);
    expect(store.getters['firewallrules/getNumberFirewalls']).toEqual(0);
  });
});
