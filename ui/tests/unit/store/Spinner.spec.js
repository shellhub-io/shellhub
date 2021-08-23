import store from '@/store';

describe('Spinner', () => {
  const status = false;

  it('Return security default variables', () => {
    expect(store.getters['spinner/getStatus']).toEqual(status);
  });
  it('Verify changed status state in setStatus mutation', () => {
    store.commit('boxs/setStatus', !status);
    expect(store.getters['boxs/getStatus']).toEqual(!status);

    store.commit('boxs/setStatus', status);
    expect(store.getters['boxs/getStatus']).toEqual(status);
  });
});
