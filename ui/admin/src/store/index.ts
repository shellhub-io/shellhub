import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";
import { stats, StatsState } from "./modules/stats";

import { UsersState, users } from "./modules/users";
import { AuthState, auth } from "./modules/auth";
import { DevicesState, devices } from "./modules/devices";
import { firewallRules, FirewallRulesState } from "./modules/firewall_rules";
import { namespaces, NamespacesState } from "./modules/namespaces";
import { layout, LayoutState } from "./modules/layout";
import { license, LicenseState } from "./modules/license";
import { publicKeys, PublicKeysState } from "./modules/public_keys";
import { sessions, SessionsState } from "./modules/sessions";
import { snackbar, SnackbarState } from "./modules/snackbar";
import { spinner, SpinnerState } from "./modules/spinner";
import { announcement, AnnouncementState } from "./modules/announcement";
import { instance, InstanceState } from "./modules/instance";

export interface State {
  user: UsersState;
  auth: AuthState;
  devices: DevicesState;
  firewallRules: FirewallRulesState;
  layout: LayoutState;
  license: LicenseState;
  namespaces: NamespacesState;
  publicKeys: PublicKeysState;
  sessions: SessionsState;
  snackbar: SnackbarState;
  spinner: SpinnerState;
  stats: StatsState;
  announcement: AnnouncementState;
  instance: InstanceState;
}

// eslint-disable-next-line symbol-description
export const key: InjectionKey<Store<State>> = Symbol();

export const store = createStore<State>({
  modules: {
    users,
    auth,
    devices,
    firewallRules,
    layout,
    license,
    namespaces,
    publicKeys,
    sessions,
    snackbar,
    spinner,
    stats,
    announcement,
    instance,
  },
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
