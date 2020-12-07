import store from '@/store';

describe('PrivateKeys', () => {
  const numberPrivateKeys = 2;
  const privateKeys = [
    {
      name: 'shellhub',
      data: 'BBGVvbmFyZG8=',
    },
    {
      name: 'shellhub',
      data: 'AbGVvbmFyZG8=',
    },
  ];
  const privateKey = {
    name: 'shellhub',
    data: 'AbGVvbmFyZG7=',
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
    store.commit('privatekeys/removePrivateKey', privateKey.data);
    privateKeys.pop(privateKey);

    expect(store.getters['privatekeys/list']).toEqual(privateKeys);
    expect(store.getters['privatekeys/list'].length).toEqual(numberPrivateKeys);
  });
});
