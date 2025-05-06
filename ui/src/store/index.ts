import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { auth, AuthState } from "./modules/auth";
import { apiKeys, ApiKeysState } from "./modules/api_keys";
import { layout, LayoutState } from "./modules/layout";
import { users, UsersState } from "./modules/users";
import { tags, TagsState } from "./modules/tags";
import { stats, StatsState } from "./modules/stats";
import { support, SupportState } from "./modules/support";
import { spinner, SpinnerState } from "./modules/spinner";
import { snackbar, SnackbarState } from "./modules/snackbar";
import { sessions, SessionsState } from "./modules/sessions";
import { sessionRecording, SessionRecordingState } from "./modules/session_recording";
import { publicKeys, PublicKeysState } from "./modules/public_keys";
import { privateKey, PrivateKeyState } from "./modules/private_key";
import { notifications, NotificationsState } from "./modules/notifications";
import { firewallRules, FirewallRulesState } from "./modules/firewall_rules";
import { devices, DevicesState } from "./modules/devices";
import { container, ContainerState } from "./modules/container";
import { box, BoxState } from "./modules/box";
import { namespaces, NamespacesState } from "./modules/namespaces";
import { tunnels, TunnelsState } from "./modules/tunnels";
import { billing } from "./modules/billing";
import { customer, CustomerState } from "./modules/customer";
import { announcement, AnnouncementState } from "./modules/announcement";
import { connectors, ConnectorState } from "./modules/connectors";
import apiPlugin from "./plugins/api";

export interface State {
  auth: AuthState;
  apiKeys: ApiKeysState;
  billing: NamespacesState;
  box: BoxState;
  customer: CustomerState;
  connectors: ConnectorState;
  devices: DevicesState;
  container: ContainerState;
  firewallRules: FirewallRulesState;
  layout: LayoutState;
  namespaces: NamespacesState;
  tunnels: TunnelsState;
  notifications: NotificationsState;
  privateKey: PrivateKeyState;
  publicKeys: PublicKeysState;
  sessionRecording: SessionRecordingState;
  sessions: SessionsState;
  snackbar: SnackbarState;
  spinner: SpinnerState;
  stats: StatsState;
  support: SupportState;
  tags: TagsState;
  users: UsersState;
  announcement: AnnouncementState;
}

export const key: InjectionKey<Store<State>> = Symbol("store");

export const store = createStore<State>({
  modules: {
    auth,
    apiKeys,
    billing,
    box,
    connectors,
    container,
    customer,
    devices,
    firewallRules,
    layout,
    namespaces,
    tunnels,
    notifications,
    privateKey,
    publicKeys,
    sessionRecording,
    sessions,
    snackbar,
    spinner,
    stats,
    support,
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
