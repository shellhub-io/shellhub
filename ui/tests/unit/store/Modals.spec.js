import store from '@/store';

describe('Modals', () => {
  it('returns terminal', () => {
    const actual = store.getters['modals/terminal'];
    expect(actual).toEqual('');
  });
  it('return addDevice', () => {
    const actual = store.getters['modals/addDevice'];
    expect(actual).toEqual(false);
  });
  it('complete test', () => {
    const terminal = '';
    const addDevice = true;

    store.commit('modals/setTerminal', terminal);
    store.commit('modals/setAddDevice', addDevice);

    expect(store.getters['modals/terminal']).toEqual(terminal);
    expect(store.getters['modals/addDevice']).toEqual(addDevice);
  });
});
