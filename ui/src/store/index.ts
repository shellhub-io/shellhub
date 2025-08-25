import { createStore, useStore as vuexUseStore } from "vuex";

export const key = Symbol("store");

export const store = createStore({
  modules: {
  },
});

export function useStore() {
  return vuexUseStore(key);
}
