import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { auth, AuthState } from "./modules/auth";
import { layout, LayoutState } from "./modules/layout";
import { users, UsersState } from "./modules/users";
import { tags, TagsState } from "./modules/tags";
import { stats, StatsState } from "./modules/stats";
import { spinner, SpinnerState } from "./modules/spinner";
import { snackbar, SnackbarState } from "./modules/snackbar";
import { sessions, SessionsState } from "./modules/sessions";
import { security, SecurityState } from "./modules/security";
import { publicKeys, PublicKeysState } from "./modules/public_keys";
import { privateKey, PrivateKeyState } from "./modules/private_key";
import { notifications, NotificationsState } from "./modules/notifications";
import { modal, ModalState } from "./modules/modal";
import { firewallRules, FirewallRulesState } from "./modules/firewall_rules";
import { devices, DevicesState } from "./modules/devices";
import { box, BoxState } from "./modules/box";
import { namespaces, NamespacesState } from "./modules/namespaces";
import { billing } from "./modules/billing";
import { customer, CustomerState } from "./modules/customer";
import { announcement, AnnouncementState } from "./modules/announcement";
import apiPlugin from "./plugins/api";

export interface State {
  auth: AuthState;
  billing: NamespacesState;
  customer: CustomerState;
  box: BoxState;
  devices: DevicesState;
  firewallRules: FirewallRulesState;
  layout: LayoutState;
  modal: ModalState;
  namespaces: NamespacesState;
  notifications: NotificationsState;
  privateKey: PrivateKeyState;
  publicKeys: PublicKeysState;
  security: SecurityState;
  sessions: SessionsState;
  snackbar: SnackbarState;
  spinner: SpinnerState;
  stats: StatsState;
  tags: TagsState;
  users: UsersState;
  announcement: AnnouncementState;
}

export const key: InjectionKey<Store<State>> = Symbol("store");

export const store = createStore<State>({
  modules: {
    auth,
    customer,
    billing,
    box,
    devices,
    firewallRules,
    layout,
    modal,
    namespaces,
    notifications,
    privateKey,
    publicKeys,
    security,
    sessions,
    snackbar,
    spinner,
    stats,
    tags,
    users,
    announcement,
  },
  plugins: [
    apiPlugin,
  ],
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
