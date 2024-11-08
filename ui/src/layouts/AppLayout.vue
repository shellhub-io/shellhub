<template>
  <v-app
    :theme="getStatusDarkMode"
    v-bind="$attrs"
  >
    <v-navigation-drawer
      theme="dark"
      v-model="showNavigationDrawer"
      :permanent="permanent"
      absolute
      app
      class="bg-v-theme-surface"
      data-test="navigation-drawer"
    >
      <v-toolbar class="bg-v-theme-surface" data-test="drawer-toolbar">
        <v-spacer />
        <router-link
          to="/"
          class="text-decoration-none"
        >
          <v-img
            :src="Logo"
            min-width="140"
            alt="Shell logo, a cloud with the writing 'ShellHub' on the right side"
            data-test="logo"
          />
        </router-link>
        <v-spacer />
      </v-toolbar>

      <div class="pa-2" v-if="hasNamespaces">
        <Namespace data-test="namespace-component" />
      </div>

      <div class="d-flex justify-center" v-else-if="envVariables.isCloud">
        <v-btn
          color="primary"
          @click="showNamespaceAdd = true"
          data-test="save-btn">
          Add Namespace
        </v-btn>
        <NamespaceAdd
          v-model="showNamespaceAdd"
          enableSwitchIn
          data-test="namespaceAdd-component"
        />
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
        <v-col class="d-flex align-end justify-center">
          <QuickConnection />
        </v-col>

      </v-list>
    </v-navigation-drawer>

    <SnackbarComponent />

    <AppBar v-model="showNavigationDrawer" data-test="app-bar" />

    <v-main data-test="main">
      <slot>
        <v-container
          class="pa-8"
          fluid
          data-test="container"
        >
          <router-view :key="currentRoute.value.path" />
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
import NamespaceAdd from "@/components/Namespace/NamespaceAdd.vue";

const router = useRouter();
const store = useStore();
const currentRoute = computed(() => router.currentRoute);
const showNamespaceAdd = ref(false);
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const getStatusDarkMode = computed(
  () => store.getters["layout/getStatusDarkMode"],
);

const { lgAndUp } = useDisplay();

const permanent = computed(() => lgAndUp.value);
const showNavigationDrawer = ref(lgAndUp.value);

const hasSpinner = computed({
  get() { return store.getters["spinner/status"]; },
  set(v) { store.dispatch("spinner/setStatus", v); },
});

onMounted(() => {
  store.dispatch("privateKey/fetch");
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Settings";
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
