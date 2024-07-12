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
        <v-divider class="ma-2" />
      </v-app-bar-title>

      <div class="pa-2">
        <Namespace data-test="namespace-component" />
        <v-divider class="ma-2" />
      </div>

      <v-list class="bg-v-theme-surface" data-test="list">
        <v-list-item
          v-for="item in items"
          :key="item.title"
          :to="item.path"
          lines="two"
          class="mb-2"
          :disabled="disableItem(item.title)"
          data-test="list-item"
        >
          <div class="d-flex align-center">
            <div class="mr-3">
              <v-icon data-test="icon">
                {{ item.icon }}
              </v-icon>
            </div>

            <v-list-item-title :data-test="item.icon + '-listItem'">
              {{ item.title }}
              <v-chip
                v-if="!envVariables.isCloud && !envVariables.isEnterprise && envVariables.premiumPaywall && item.isPremium"
                density="comfortable"
                label
                variant="outlined"
                size="x-small"
                class="ml-1"
                color="yellow"
                prepend-icon="mdi-crown">Premium</v-chip>
            </v-list-item-title>
          </div>
        </v-list-item>
        <v-col class="d-flex align-end justify-center">
          <NewConnection />
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
import NewConnection from "../components/NewConnection/NewConnection.vue";

const items = [
  {
    icon: "mdi-view-dashboard",
    title: "Dashboard",
    path: "/",
  },
  {
    icon: "mdi-cellphone-link",
    title: "Devices",
    path: "/devices",
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

const router = useRouter();
const store = useStore();
const currentRoute = computed(() => router.currentRoute);
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const getStatusDarkMode = computed(
  () => store.getters["layout/getStatusDarkMode"],
);

const { lgAndUp } = useDisplay();

const showNavigationDrawer = ref(lgAndUp);

const hasSpinner = computed({
  get() { return store.getters["spinner/status"]; },
  set(v) { store.dispatch("spinner/setStatus", v); },
});

onMounted(() => {
  store.dispatch("privateKey/fetch");
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Dashboard";

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
