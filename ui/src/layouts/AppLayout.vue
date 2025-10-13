<template>
  <v-navigation-drawer
    :theme="theme"
    v-model="showNavigationDrawer"
    :permanent="permanent"
    absolute
    app
    class="bg-v-theme-surface"
    data-test="navigation-drawer"

  >
    <v-toolbar class="bg-v-theme-surface border-b-thin" data-test="drawer-toolbar">
      <div class="w-100 d-flex align-center justify-center">
        <router-link
          to="/"
          class="text-decoration-none"
        >
          <v-img
            :src="Logo"
            width="160"
            data-test="logo"
          />
        </router-link>
      </div>
    </v-toolbar>

    <div class="d-flex justify-center" v-if="!hasNamespaces">
      <v-btn
        color="primary"
        @click="showNamespaceAdd = true"
        data-test="save-btn">
        Add Namespace
      </v-btn>
      <NamespaceAdd
        v-model="showNamespaceAdd"
        enableSwitchIn
        data-test="namespace-add-component"
      />
    </div>

    <v-list density="compact" class="bg-v-theme-surface" data-test="list">
      <template v-for="item in visibleItems" :key="item.title">
        <v-list-group
          v-if="item.children && getFilteredChildren(item.children).length > 0"
          prepend-icon="mdi-chevron-down"
          v-model="subMenuState[item.title]"
          data-test="list-group"
        >
          <template v-slot:activator="{ props }">
            <v-list-item
              lines="two"
              v-bind="props"
              :disabled="disableItem(item.title)">
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
            v-for="child in getFilteredChildren(item.children)"
            :key="child.title"
            :to="child.path"
            :disabled="disableItem(item.title)"
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
            <v-chip
              v-if="item.isBeta && envVariables.isCloud"
              label
              color="yellow"
              size="x-small"
              data-test="isBeta-chip"
            >
              BETA
            </v-chip>

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

  <Snackbar />

  <AppBar v-model="showNavigationDrawer" data-test="app-bar" />

  <v-main data-test="main">
    <slot>
      <v-container
        :class="{ 'pa-8': true, 'container-light-bg': theme === 'light' }"
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

  <UserWarning data-test="userWarning-component" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useDisplay } from "vuetify";
import Logo from "../assets/logo-inverted.png";
import { envVariables } from "../envVariables";
import UserWarning from "../components/User/UserWarning.vue";
import AppBar from "../components/AppBar/AppBar.vue";
import QuickConnection from "../components/QuickConnection/QuickConnection.vue";
import NamespaceAdd from "@/components/Namespace/NamespaceAdd.vue";
import Snackbar from "@/components/Snackbar/Snackbar.vue";
import useLayoutStore from "@/store/modules/layout";
import useNamespacesStore from "@/store/modules/namespaces";
import usePrivateKeysStore from "@/store/modules/private_keys";
import useSpinnerStore from "@/store/modules/spinner";

defineOptions({
  inheritAttrs: false,
});

const router = useRouter();
const layoutStore = useLayoutStore();
const namespacesStore = useNamespacesStore();
const spinnerStore = useSpinnerStore();
const { getPrivateKeyList } = usePrivateKeysStore();
const currentRoute = computed(() => router.currentRoute);
const showNamespaceAdd = ref(false);
const hasNamespaces = computed(() => namespacesStore.namespaceList.length !== 0);
const theme = computed(() => layoutStore.theme);

const { lgAndUp } = useDisplay();

const permanent = computed(() => lgAndUp.value);
const showNavigationDrawer = ref(lgAndUp.value);

const hasSpinner = computed({
  get() { return spinnerStore.status; },
  set(newStatus) { spinnerStore.status = newStatus; },
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Settings";

const isItemHidden = (meta?: Record<string, unknown>) => {
  if (meta?.isHidden && typeof meta.isHidden === "function") return meta.isHidden();
  return false;
};

const items = computed(() => {
  const routes = router.getRoutes();

  // Get all parent routes that should show in sidebar
  const parentRoutes = routes.filter((route) => route.meta?.showInSidebar && route.meta?.sidebarOrder);

  return parentRoutes
    .map((route) => ({
      icon: route.meta?.icon as string,
      title: route.meta?.title as string,
      path: route.path,
      isPremium: route.meta?.isPremium as boolean,
      isBeta: route.meta?.isBeta as boolean,
      hidden: isItemHidden(route.meta),
      sidebarOrder: route.meta?.sidebarOrder as number,
      children: route.children
        ?.filter((child) => child.meta?.showInSidebar)
        .map((child) => ({
          title: child.meta?.title as string,
          path: child.path.startsWith("/") ? child.path : `${route.path}/${child.path}`,
          hidden: isItemHidden(child.meta),
        })),
    }))
    .sort((a, b) => a.sidebarOrder - b.sidebarOrder);
});

const subMenuState = reactive({});

items.value.forEach((item) => {
  if (item.children) {
    subMenuState[item.title] = false;
  }
});

function getFilteredChildren(children) {
  return children.filter((child) => !child.hidden);
}

const visibleItems = computed(() => items.value.filter((item) => !item.hidden));

onMounted(() => { getPrivateKeyList(); });

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

.container-light-bg {
  background-image: linear-gradient(155deg, rgb(var(--v-theme-primary),0.10) 0%, transparent 0%), url(/bg-inverted.svg);
}
</style>
