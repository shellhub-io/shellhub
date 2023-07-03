<template>
  <v-app
    :theme="getStatusDarkMode"
    v-bind="$attrs"
  >
    <!-- <v-row>
      <v-col class="d-flex flex-column align-center">
        <p>dsaasd</p>
      </v-col>
    </v-row> -->
    <v-navigation-drawer
      theme="dark"
      v-model="showNavigationDrawer"
      :permanent="lgAndUp"
      absolute
      app
      class="bg-v-theme-surface"
    >
      <v-app-bar-title>
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
            />
          </div>
        </router-link>
        <v-divider class="ma-2" />
      </v-app-bar-title>

      <div class="pa-2">
        <Namespace data-test="namespace-component" />
        <v-divider class="ma-2" />
      </div>

      <v-list class="bg-v-theme-surface">
        <v-list-item
          v-for="item in visibleItems"
          :key="item.title"
          :to="item.path"
          lines="two"
          class="mb-2"
          :disabled="disableItem(item.title)"
        >
          <div class="d-flex align-center">
            <div class="mr-3">
              <v-icon>
                {{ item.icon }}
              </v-icon>
            </div>

            <v-list-item-title :data-test="item.icon + '-listItem'">
              {{ item.title }}
            </v-list-item-title>
          </div>
        </v-list-item>
        <v-col class="d-flex align-end justify-center">
          <NewConnection />
        </v-col>

      </v-list>
    </v-navigation-drawer>

    <SnackbarComponent />

    <AppBar v-model="showNavigationDrawer" />

    <v-main>
      <slot>
        <v-container
          class="pa-8"
          fluid
        >
          <router-view :key="currentRoute.value.path" />
        </v-container>
      </slot>
    </v-main>

    <v-overlay
      :scrim="false"
      disabled
      v-model="hasSpinner"
    >
      <v-col class="full-width-height d-flex justify-center align-center">
        <v-progress-circular
          indeterminate
          size="64"
          alt="Request loading"
        />
      </v-col>
    </v-overlay>
  </v-app>

  <UserWarning data-test="userWarning-component" />
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useDisplay } from "vuetify";
import Logo from "../assets/logo-inverted.png";
import { useStore } from "../store";
import UserWarning from "../components/User/UserWarning.vue";
import Namespace from "../../src/components/Namespace/Namespace.vue";
import AppBar from "../components/AppBar/AppBar.vue";
import { envVariables } from "../envVariables";
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
    hidden: !envVariables.isEnterprise,
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
const visibleItems = computed(() => items.filter((item) => !item.hidden));
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const getStatusDarkMode = computed(
  () => store.getters["layout/getStatusDarkMode"],
);

const { lgAndUp } = useDisplay();
const showNavigationDrawer = ref(lgAndUp);
const hasSpinner = computed(() => store.getters["spinner/status"]);

onMounted(() => {
  store.dispatch("privateKey/fetch");
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Dashboard";
</script>

<style lang="css" scoped>
.full-width-height {
  width: 100vw !important;
  height: 100vh !important;
}
</style>
