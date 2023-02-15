import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Spinner', () => {
  const status = false;

  it('Return spinner default variables', () => {
    expect(store.getters['spinner/status']).toEqual(status);
  });
  it('Verify changed status state in setStatus mutation', () => {
    store.commit('spinner/setStatus', !status);
    expect(store.getters['spinner/status']).toEqual(!status);

    store.commit('spinner/setStatus', status);
    expect(store.getters['spinner/status']).toEqual(status);
  });
});
