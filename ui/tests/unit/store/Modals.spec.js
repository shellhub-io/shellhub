import store from '@/store';

describe('Modals', () => {
  const terminal = '';
  const addDevice = true;

  it('Return modal default variables', () => {
    expect(store.getters['modals/terminal']).toEqual('');
    expect(store.getters['modals/addDevice']).toEqual(false);
  });
  it('Verify initial states change for mutation setTerminal', () => {
    store.commit('modals/setTerminal', terminal);

    expect(store.getters['modals/terminal']).toEqual(terminal);
  });
  it('Verify initial states change for mutation addDevice', () => {
    store.commit('modals/setAddDevice', addDevice);

    expect(store.getters['modals/addDevice']).toEqual(addDevice);
  });
});
