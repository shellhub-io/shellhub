import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Snackbar', () => {
  const sessionRecord = true;

  it('Return security default variables', () => {
    expect(store.getters['security/get']).toEqual(sessionRecord);
  });

  it('Verify set security state for setSecurity mutation', () => {
    store.commit("security/setSecurity", !sessionRecord)
    expect(store.getters['security/get']).toEqual(!sessionRecord);
  });
});
