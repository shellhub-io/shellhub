<template>
  <v-app
    :theme="getStatusDarkMode"
    v-bind="$attrs"
  >
    <v-navigation-drawer
      theme="dark"
      v-model="showNavigationDrawer"
      :permanent="lgAndUp"
      absolute
      app
      class="bg-v-theme-surface"
      data-test="navigation-drawer"
    >
      <v-app-bar-title data-test="app-bar-title">
        <router-link
          to="/"
          class="text-decoration-none"
        >
          <div class="d-flex justify-center pa-4 pb-2">
            <v-img
              class="d-sm-flex hidden-sm-and-down"
              :src="Logo"
              max-width="140"
              alt="Shell logo, a cloud with the writing 'ShellHub' on the right side"
              data-test="logo"
            />
          </div>
        </router-link>
      </v-app-bar-title>

      <div class="pa-2" v-if="hasNamespaces">
        <Namespace data-test="namespace-component" />
      </div>

      <v-list density="compact" class="bg-v-theme-surface" data-test="list">
        <v-list-item
          v-for="item in visibleItems"
          :key="item.title"
          :to="item.path"
          lines="two"
          class="mb-2"
          :disabled="disableItem(item.title)"
          data-test="list-item"
        >
          <template #prepend>
            <v-icon data-test="icon">
              {{ item.icon }}
            </v-icon>
          </template>
          <template #append>
            <v-icon
              v-if="item.isPremium && envVariables.isCommunity && envVariables.premiumPaywall"
              color="yellow"
              size="x-small"
              icon="mdi-crown"
              data-test="icon"
            />
          </template>
          <v-list-item-title :data-test="item.icon + '-listItem'">
            {{ item.title }}
          </v-list-item-title>
        </v-list-item>

        <v-col v-if="terminalTokens.length > 0" class="d-flex align-end justify-center pa-2">
          <TerminalSelect @open-quick-dialog="isAddNamespaceDialogVisible = true" />
        </v-col>

        <v-col v-else class="d-flex align-end justify-center">
          <div>
            <v-btn
              @click="isAddNamespaceDialogVisible = true"
              color="primary"
              tabindex="0"
              variant="elevated"
              aria-label="Dialog Quick Connection"
              data-test="quick-connection-open-btn"
              prepend-icon="mdi-link"
            >
              Quick Connection
            </v-btn>
            <div>
              <p
                class="text-caption text-md font-weight-bold text-grey-darken-1 ma-1"
                data-test="quick-connect-instructions"
              >
                Press "Ctrl + K" to Quick Connect!
              </p>
            </div>
          </div>
        </v-col>

      </v-list>
    </v-navigation-drawer>
    <SnackbarComponent />

    <AppBar
      v-model:showNavigationDrawer="showNavigationDrawer"
      data-test="app-bar" />

    <v-main data-test="main">
      <slot>
        <v-container
          :class="noGaps ? 'd-flex fill-height pa-0' : 'pa-8'"
          fluid
          data-test="container"
        >
          <router-view v-slot="{ Component }">
            <keep-alive include="Connection">
              <component :is="Component" :key="currentRoute.value.path" />
            </keep-alive>
          </router-view>
        </v-container>
      </slot>
    </v-main>

    <v-overlay
      :model-value="hasSpinner"
      :scrim="false"
      contained
      class="align-center justify-center w-100 h-100"
      data-test="overlay"
    >
      <v-progress-circular
        indeterminate
        size="64"
        alt="Request loading"
        data-test="progress-circular"
      />
    </v-overlay>
  </v-app>

  <QuickConnection v-model="isAddNamespaceDialogVisible" />
  <UserWarning data-test="userWarning-component" />
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useDisplay } from "vuetify";
import Logo from "../assets/logo-inverted.png";
import { envVariables } from "../envVariables";
import { useStore } from "../store";
import UserWarning from "../components/User/UserWarning.vue";
import Namespace from "../../src/components/Namespace/Namespace.vue";
import AppBar from "../components/AppBar/AppBar.vue";
import QuickConnection from "../components/QuickConnection/QuickConnection.vue";
import TerminalSelect from "@/components/Terminal/TerminalSelect.vue";

const router = useRouter();
const store = useStore();
const currentRoute = computed(() => router.currentRoute);
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const getStatusDarkMode = computed(
  () => store.getters["layout/getStatusDarkMode"],
);
const isAddNamespaceDialogVisible = ref(false);
const { lgAndUp } = useDisplay();

const showNavigationDrawer = ref(lgAndUp);

const hasSpinner = computed({
  get() { return store.getters["spinner/status"]; },
  set(v) { store.dispatch("spinner/setStatus", v); },
});

onMounted(() => {
  store.dispatch("privateKey/fetch");
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Home";
const showConnector = computed(() => (envVariables.isCommunity && !envVariables.premiumPaywall) || !envVariables.hasConnector);
const showFirewall = computed(() => envVariables.isCommunity && !envVariables.premiumPaywall);
const items = [
  {
    icon: "mdi-home",
    title: "Home",
    path: "/",
  },
  {
    icon: "mdi-cellphone-link",
    title: "Devices",
    path: "/devices",
  },
  {
    icon: "mdi-server",
    title: "Containers",
    path: "/containers",
  },
  {
    icon: "mdi-docker",
    title: "Connectors",
    path: "/connectors",
    isPremium: true,
    hidden: showConnector.value,
  },
  {
    icon: "mdi-history",
    title: "Sessions",
    path: "/sessions",
  },
  {
    icon: "mdi-security",
    title: "Firewall Rules",
    path: "/firewall/rules",
    isPremium: true,
    hidden: showFirewall.value,
  },
  {
    icon: "mdi-key",
    title: "Public Keys",
    path: "/sshkeys/public-keys",
  },
  {
    icon: "mdi-cog",
    title: "Settings",
    path: "/settings",
  },
];

const visibleItems = computed(() => items.filter((item) => !item.hidden));
const terminalTokens = computed(() => Object.keys(store.getters["terminals/getTerminal"]));
const noGaps = computed(() => router.currentRoute.value.meta.noGaps);

onMounted(async () => {
  await store.dispatch("terminals/fetchThemes");
});

defineExpose({
  items,
  lgAndUp,
});
</script>

<style lang="css" scoped>
.full-width-height {
  width: 100vw !important;
  height: 100vh !important;
}
</style>
