import store from '@/store';

describe('Namespace', () => {
  const owner = true;
  const numberNamespaces = 4;

  const namespaces = [
    {
      name: 'namespace1',
      owner: 'user1',
      members: [{ name: 'user3' }, { name: 'user4' }, { name: 'user5' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
    },
    {
      name: 'namespace2',
      owner: 'user1',
      members: [{ name: 'user3' }, { name: 'user4' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484714',
    },
    {
      name: 'namespace3',
      owner: 'user1',
      members: [{ name: 'user6' }, { name: 'user 7' }, { name: 'user 8' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    },
    {
      name: 'namespace4',
      owner: 'user1',
      members: [{ name: 'user6' }, { name: 'user7' }],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484716',
    },
  ];

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    members: [{ name: 'user6' }, { name: 'user7' }, { name: 'user8' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    settings: {
      webhook: {
        url: 'http://192.178.90.2:8080',
        active: true,
      },
    },
  };

  it('Returns namespaces default variables', () => {
    expect(store.getters['namespaces/list']).toEqual([]);
    expect(store.getters['namespaces/get']).toEqual({});
    expect(store.getters['namespaces/getNumberNamespaces']).toEqual(0);
    expect(store.getters['namespaces/owner']).toEqual(!owner);
  });
  // mutations tests
  it('Verify initial state change for setNamespaces mutation', () => {
    store.commit('namespaces/setNamespaces', { data: namespaces, headers: { 'x-total-count': numberNamespaces } });
    expect(store.getters['namespaces/list']).toEqual(namespaces);
    expect(store.getters['namespaces/getNumberNamespaces']).toEqual(numberNamespaces);
  });
  it('Verify initial state change for setWebhook mutation', () => {
    store.commit('namespaces/setWebhook', namespace.settings.webhook);
    expect(store.getters['namespaces/webhookUrl']).toEqual(namespace.settings.webhook.url);
  });
  it('Verify initial state change for setWebhookStatus mutation', () => {
    store.commit('namespaces/setWebhookStatus', namespace.settings.webhook.active);
    expect(store.getters['namespaces/webhookActive']).toEqual(namespace.settings.webhook.active);
  });
  it('Verify initial state change for setNamespace mutation', () => {
    const { active, url } = namespace.settings.webhook;

    store.commit('namespaces/setNamespace', { data: namespace });
    expect(store.getters['namespaces/get']).toEqual(namespace);
    expect(store.getters['namespaces/webhookUrl']).toEqual(url);
    expect(store.getters['namespaces/webhookActive']).toEqual(active);
  });
  it('Verify changed member list for removeMember mutation', () => {
    const lengthListOfMembers = namespace.members.length;
    const lastUsername = namespace.members[lengthListOfMembers - 1];

    store.commit('namespaces/removeMember', lastUsername);
    expect(store.getters['namespaces/get'].members).toHaveLength(lengthListOfMembers - 1);
    store.getters['namespaces/get'].members.forEach((member) => {
      expect(member.name === lastUsername).toBeFalsy();
    });
  });
  it('Verify changed namespaces list for removeNamespace mutation', () => {
    store.commit('namespaces/removeNamespace', namespace.tenant_id);
    expect(store.getters['namespaces/list']).toHaveLength(numberNamespaces - 1);
    store.getters['namespaces/list'].forEach((item) => {
      expect(item.tenant_id === namespace.tenant_id).toBeFalsy();
    });
  });
  it('Verify changed owner for setOwnerStatus mutation', () => {
    store.commit('namespaces/setOwnerStatus', owner);
    expect(store.getters['namespaces/owner']).toEqual(owner);
  });
  it('Clears the namespace variables from store', () => {
    store.commit('namespaces/clearNamespaceList');
    store.commit('namespaces/clearObjectNamespace');
    expect(store.getters['namespaces/list']).toEqual([]);
    expect(store.getters['namespaces/get']).toEqual({});
  });
});
