import store from '@/store';

describe('PrivateKeys', () => {
  const numberPrivateKeys = 3;

  const privateKeys = [
    {
      name: 'key1',
      data: 'BBGVvbmFyZG8=',
    },
    {
      name: 'key2',
      data: 'AbGVvbmFyZG8=',
    },
    {
      name: 'key3',
      data: 'CbGVvbmFyZG8=',
    },
  ];
  const privateKey = {
    name: 'key4',
    data: 'AbGVvbmFyZG7=',
  };
  const privateKey2 = {
    name: 'key2',
    data: 'AbGVvbmFyZG8=',
  };

  it('Return private key default variables', () => {
    expect(store.getters['privatekeys/list']).toEqual([]);
    expect(store.getters['privatekeys/getNumberPrivateKeys']).toEqual(0);
  });

  it('Verify initial state change for setPrivateKey mutation', () => {
    jest.spyOn(Storage.prototype, 'getItem').mockReturnValue(privateKeys);
    privateKeys.forEach((pk) => {
      store.commit('privatekeys/setPrivateKey', pk);
    });

    expect(store.getters['privatekeys/list']).toEqual(privateKeys);
    expect(store.getters['privatekeys/getNumberPrivateKeys']).toEqual(numberPrivateKeys);
  });
  it('Verify inital state change for setPrivateKey mutation', () => {
    store.commit('privatekeys/setPrivateKey', privateKey);
    privateKeys.push(privateKey);

    expect(store.getters['privatekeys/list']).toEqual(privateKeys);
    expect(store.getters['privatekeys/getNumberPrivateKeys']).toEqual(numberPrivateKeys + 1);
  });
  it('Verify remove private key item from list for removePrivateKey mutation', () => {
    const currentPrivateKeys = store.getters['privatekeys/list'];
    const currentNumberPrivateKeys = store.getters['privatekeys/getNumberPrivateKeys'];

    store.commit('privatekeys/removePrivateKey', privateKey2.data);
    expect(store.getters['privatekeys/list']).toEqual(currentPrivateKeys.filter((pk) => pk.data !== privateKey2.data));
    expect(store.getters['privatekeys/getNumberPrivateKeys']).toEqual(currentNumberPrivateKeys - 1);
  });
});
