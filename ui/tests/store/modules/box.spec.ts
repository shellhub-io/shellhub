import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Snackbar', () => {
  const status = true;

  it('Return security default variables', () => {
    expect(store.getters['box/getStatus']).toEqual(status);
  });
  it('Verify changed status state in setStatus mutation', () => {
    store.commit('box/setStatus', !status);
    expect(store.getters['box/getStatus']).toEqual(!status);

    store.commit('box/setStatus', status);
    expect(store.getters['box/getStatus']).toEqual(status);
  });
});
