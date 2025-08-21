import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { users, UsersState } from "./modules/users";
import { tags, TagsState } from "./modules/tags";
import { stats, StatsState } from "./modules/stats";
import { support, SupportState } from "./modules/support";
import { webEndpoints, WebEndpointsState } from "./modules/web_endpoints";

export interface State {
  webEndpoints: WebEndpointsState;
  stats: StatsState;
  support: SupportState;
  tags: TagsState;
  users: UsersState;
}

export const key: InjectionKey<Store<State>> = Symbol("store");

export const store = createStore<State>({
  modules: {
    webEndpoints,
    stats,
    support,
    tags,
    users,
  },
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
