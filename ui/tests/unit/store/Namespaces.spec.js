import store from '@/store';

describe('Namespace', () => {
  const numberNamespaces = 4;
  const namespaces = [
    {
      name: 'namespace1',
      owner: 'user1',
      member_names: ['user3', 'user4', 'user5'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484713',
    },
    {
      name: 'namespace2',
      owner: 'user1',
      member_names: ['user3', 'user4'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484714',
    },
    {
      name: 'namespace3',
      owner: 'user1',
      member_names: ['user6', 'user7', 'user8'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    },
    {
      name: 'namespace4',
      owner: 'user1',
      member_names: ['user6', 'user7'],
      tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484716',
    },
  ];

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
  };

  it('Returns namespaces default variables', () => {
    expect(store.getters['namespaces/list']).toEqual([]);
    expect(store.getters['namespaces/get']).toEqual({});
    expect(store.getters['namespaces/getNumberNamespaces']).toEqual(0);
  });
  // mutations tests
  it('Verify initial state change for setNamespaces mutation', () => {
    store.commit('namespaces/setNamespaces', { data: namespaces, headers: { 'x-total-count': numberNamespaces } });
    expect(store.getters['namespaces/list']).toEqual(namespaces);
    expect(store.getters['namespaces/getNumberNamespaces']).toEqual(numberNamespaces);
  });
  it('Verify initial state change for setNamespace mutation', () => {
    store.commit('namespaces/setNamespace', { data: namespace });
    expect(store.getters['namespaces/get']).toEqual(namespace);
  });
  it('Verify changed member list for removeMember mutation', () => {
    const lengthListOfMembers = namespace.member_names.length;
    const lastUsername = namespace.member_names[lengthListOfMembers - 1];
    store.commit('namespaces/removeMember', lastUsername);
    expect(store.getters['namespaces/get'].member_names).toHaveLength(lengthListOfMembers - 1);
    store.getters['namespaces/get'].member_names.forEach((member) => {
      expect(member === lastUsername).toBeFalsy();
    });
  });
  it('Verify changed namespaces list for removeNamespace mutation', () => {
    store.commit('namespaces/removeNamespace', namespace.tenant_id);
    expect(store.getters['namespaces/list']).toHaveLength(numberNamespaces - 1);
    store.getters['namespaces/list'].forEach((item) => {
      expect(item.tenant_id === namespace.tenant_id).toBeFalsy();
    });
  });
});
