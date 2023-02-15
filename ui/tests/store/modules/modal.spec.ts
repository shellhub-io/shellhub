import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Modals', () => {
  const terminal = '';
  const addDevice = true;

  it('Return modal default variables', () => {
    expect(store.getters['modal/terminal']).toEqual('');
    expect(store.getters['modal/addDevice']).toEqual(false);
  });
  it('Verify initial states change for mutation setTerminal', () => {
    store.commit('modal/setTerminal', terminal);
    expect(store.getters['modal/terminal']).toEqual(terminal);
  });
  it('Verify initial states change for mutation addDevice', () => {
    store.commit('modal/setAddDevice', addDevice);
    expect(store.getters['modal/addDevice']).toEqual(addDevice);
  });
});
