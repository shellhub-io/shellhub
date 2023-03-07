import { InjectionKey } from "vue";
import { createStore, Store, useStore as vuexUseStore } from "vuex";

import { createAuthModule, AuthState } from "./modules/auth";
import { createLayoutModule, LayoutState } from "./modules/layout";
import { createUsersModule, UsersState } from "./modules/users";
import { createTagsModule, TagsState } from "./modules/tags";
import { createStatsModule, StatsState } from "./modules/stats";
import { createSpinnerModule, SpinnerState } from "./modules/spinner";
import { createSnackbarModule, SnackbarState } from "./modules/snackbar";
import { createSessionsModule, SessionsState } from "./modules/sessions";
import { createSecurityModule, SecurityState } from "./modules/security";
import { createPublicKeysModule, PublicKeysState } from "./modules/public_keys";
import { createPrivateKeyModule, PrivateKeyState } from "./modules/private_key";
import { createNotificationsModule, NotificationsState } from "./modules/notifications";
import { createModalModule, ModalState } from "./modules/modal";
import { createMobileModule, MobileState } from "./modules/mobile";
import { createFirewallRulesModule, FirewallRulesState } from "./modules/firewall_rules";
import { createDeviceModule, DevicesState } from "./modules/devices";
import { createBoxModule, BoxState } from "./modules/box";
import { createNamespacesModule, NamespacesState } from "./modules/namespaces";
import { createBillingModule } from "./modules/billing";
import { AnnouncementState, createAnnouncementModule } from "./modules/announcement";

export interface State {
  auth: AuthState;
  billing: NamespacesState;
  box: BoxState;
  devices: DevicesState;
  firewallRules: FirewallRulesState;
  layout: LayoutState;
  mobile: MobileState;
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
    auth: createAuthModule(),
    billing: createBillingModule(),
    box: createBoxModule(),
    devices: createDeviceModule(),
    firewallRules: createFirewallRulesModule(),
    layout: createLayoutModule(),
    mobile: createMobileModule(),
    modal: createModalModule(),
    namespaces: createNamespacesModule(),
    notifications: createNotificationsModule(),
    privateKey: createPrivateKeyModule(),
    publicKeys: createPublicKeysModule(),
    security: createSecurityModule(),
    sessions: createSessionsModule(),
    snackbar: createSnackbarModule(),
    spinner: createSpinnerModule(),
    stats: createStatsModule(),
    tags: createTagsModule(),
    users: createUsersModule(),
    announcement: createAnnouncementModule(),
  },
});

export function useStore(): Store<State> {
  return vuexUseStore(key);
}
