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
        <template v-for="item in visibleItems" :key="item.title">
          <v-list-group
            v-if="item.children"
            prepend-icon="mdi-chevron-down"
            v-model="subMenuState[item.title]"
            data-test="list-group"
          >
            <template v-slot:activator="{ props }">
              <v-list-item lines="two" v-bind="props">
                <template #prepend>
                  <v-icon data-test="icon">
                    {{ item.icon }}
                  </v-icon>
                </template>
                <v-list-item-title>
                  {{ item.title }}
                </v-list-item-title>
              </v-list-item>
            </template>

            <v-list-item
              v-for="child in item.children"
              :key="child.title"
              :to="child.path"
              data-test="list-item"
            >
              <v-list-item-title :data-test="child.title + '-listItem'">
                {{ child.title }}
              </v-list-item-title>
            </v-list-item>
          </v-list-group>

          <v-list-item
            v-else
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
        </template>

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
import { computed, reactive, ref, onMounted } from "vue";
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
const namespacedInstance = computed(() => localStorage.getItem("tenant") !== "");
const hasNamespace = computed(() => store.getters["namespaces/getNumberNamespaces"] !== 0);

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
    icon: "mdi-account-group",
    title: "Team",
    path: "/team",
    children: [
      {
        title: "Members",
        path: "/team/members",
      },
      {
        title: "API Keys",
        path: "/team/api-keys",
      },
    ],
  },
  {
    icon: "mdi-cog",
    title: "Settings",
    path: "/settings",
    children: [
      {
        title: "Profile",
        path: "/settings/profile",
      },
      {
        title: "Namespace",
        path: "/settings/namespace",
        hidden: !namespacedInstance.value,
      },
      {
        title: "Private Keys",
        path: "/settings/private-keys",
      },
      {
        title: "Tags",
        path: "/settings/tags",
        hidden: !namespacedInstance.value,
      },
      {
        title: "Billing",
        path: "/settings/billing",
        hidden: !(
          envVariables.billingEnable
          && envVariables.isCloud
          && hasNamespace.value
        ),
      },
    ],
  },
];

const subMenuState = reactive({});

items.forEach((item) => {
  if (item.children) {
    subMenuState[item.title] = false;
  }
});

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

.v-container {
  min-height: calc(100vh - 64px);
  background-image: linear-gradient(155deg, rgb(var(--v-theme-primary),0.10) 0%, transparent 30%), url(/bg.svg);
  background-position: 0% 0;
  background-repeat: no-repeat;
  background-size: auto;
}
</style>
