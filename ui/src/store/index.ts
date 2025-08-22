import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { webEndpoints, WebEndpointsState } from "./modules/web_endpoints";

export interface State {
  webEndpoints: WebEndpointsState;
}

export const key: InjectionKey<Store<State>> = Symbol("store");

export const store = createStore<State>({
  modules: {
    webEndpoints,
  },
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
