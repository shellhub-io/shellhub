import store from '@/store';

describe('Snackbar', () => {
  const sessionRecord = true;

  it('Return security default variables', () => {
    expect(store.getters['security/get']).toEqual(sessionRecord);
  });
});
