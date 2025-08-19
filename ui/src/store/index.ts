import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { users, UsersState } from "./modules/users";
import { tags, TagsState } from "./modules/tags";
import { stats, StatsState } from "./modules/stats";
import { support, SupportState } from "./modules/support";
import { spinner, SpinnerState } from "./modules/spinner";
import { sessions, SessionsState } from "./modules/sessions";
import { sessionRecording, SessionRecordingState } from "./modules/session_recording";
import { publicKeys, PublicKeysState } from "./modules/public_keys";
import { privateKey, PrivateKeyState } from "./modules/private_key";
import { notifications, NotificationsState } from "./modules/notifications";
import { webEndpoints, WebEndpointsState } from "./modules/web_endpoints";

export interface State {
  webEndpoints: WebEndpointsState;
  notifications: NotificationsState;
  privateKey: PrivateKeyState;
  publicKeys: PublicKeysState;
  sessionRecording: SessionRecordingState;
  sessions: SessionsState;
  spinner: SpinnerState;
  stats: StatsState;
  support: SupportState;
  tags: TagsState;
  users: UsersState;
}

export const key: InjectionKey<Store<State>> = Symbol("store");

export const store = createStore<State>({
  modules: {
    webEndpoints,
    notifications,
    privateKey,
    publicKeys,
    sessionRecording,
    sessions,
    spinner,
    stats,
    support,
    tags,
    users,
  },
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
