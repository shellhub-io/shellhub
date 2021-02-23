import store from '@/store';

describe('PublicKeys', () => {
  const numberPublicKeys = 2;
  const publicKeys = [
    {
      data: 'BBGVvbmFyZG8=',
      fingerprint: 'b8:26:d5',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      name: 'shellhub',
    },
    {
      data: 'AbGVvbmFyZG8=',
      fingerprint: 'b7:25:f8',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      name: 'shellhub',
    },
  ];
  const publicKey = {
    data: 'AbGVvbmFyZG8=',
    fingerprint: 'b7:25:f8',
    created_at: '2020-11-23T20:59:13.323Z',
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    name: 'shellhub',
  };
  const pagePerpageInitialValue = {
    page: 0,
    perPage: 10,
  };
  const pagePerpageValue = {
    page: 1,
    perPage: 50,
  };

  it('Return public key default variables', () => {
    expect(store.getters['publickeys/list']).toEqual([]);
    expect(store.getters['publickeys/get']).toEqual({});
    expect(store.getters['publickeys/getNumberPublicKeys']).toEqual(0);
    expect(store.getters['publickeys/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['publickeys/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });

  it('Verify initial state change for setPublicKeys mutation', () => {
    store.commit('publickeys/setPublicKeys', { data: publicKeys, headers: { 'x-total-count': numberPublicKeys } });
    expect(store.getters['publickeys/list']).toEqual(publicKeys);
    expect(store.getters['publickeys/getNumberPublicKeys']).toEqual(numberPublicKeys);
  });
  it('Verify inital state change for setPublicKey mutation', () => {
    store.commit('publickeys/setPublicKey', { data: publicKey });
    expect(store.getters['publickeys/get']).toEqual(publicKey);
  });
  it('Verify inital state change for setPagePerpage mutation', () => {
    store.commit('publickeys/setPagePerpage', pagePerpageValue);
    expect(store.getters['publickeys/getPage']).toEqual(pagePerpageValue.page);
    expect(store.getters['publickeys/getPerPage']).toEqual(pagePerpageValue.perPage);
  });
  it('Verify inital state change for resetPagePerpage mutation', () => {
    store.commit('publickeys/resetPagePerpage');
    expect(store.getters['publickeys/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['publickeys/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });
  it('Verify remove public key item from list for removePublicKey mutation', () => {
    store.commit('publickeys/removePublicKey', publicKey.id);
    expect(store.getters['publickeys/list'].length).toEqual(numberPublicKeys - 1);
  });
  it('Verify changed public key object state for clearObjectPublicKey mutation', () => {
    store.commit('publickeys/clearObjectPublicKey');
    expect(store.getters['publickeys/get']).toEqual({});
  });
  it('Verify changed firewall list state for clearListPublicKey mutation', () => {
    store.commit('publickeys/clearListPublicKeys');
    expect(store.getters['publickeys/list']).toEqual([]);
    expect(store.getters['publickeys/getNumberPublicKeys']).toEqual(0);
  });
});
