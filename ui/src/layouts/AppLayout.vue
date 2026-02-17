<template>
  <v-navigation-drawer
    v-model="showNavigationDrawer"
    :theme="theme"
    :permanent="permanent"
    app
    class="bg-v-theme-surface"
    data-test="navigation-drawer"
  >
    <v-toolbar
      class="bg-v-theme-surface border-b-thin"
      data-test="drawer-toolbar"
    >
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

    <v-list
      density="compact"
      class="bg-v-theme-surface"
      data-test="list"
    >
      <template
        v-for="item in visibleItems"
        :key="item.title"
      >
        <v-list-group
          v-if="item.children && getFilteredChildren(item.children).length > 0"
          v-model="subMenuState[item.title]"
          prepend-icon="mdi-chevron-down"
          data-test="list-group"
        >
          <template #activator="{ props }">
            <v-list-item
              lines="two"
              v-bind="props"
              :disabled="disableItem(item.title)"
            >
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
        <QuickConnection :disabled="!hasNamespaces" />
      </v-col>
    </v-list>
  </v-navigation-drawer>

  <Snackbar />

  <AppBar
    v-model="showNavigationDrawer"
    data-test="app-bar"
  />

  <v-main data-test="main">
    <slot>
      <v-container
        class="pa-8 container"
        :class="{ 'container-light-bg': theme === 'light' }"
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
import Snackbar from "@/components/Snackbar/Snackbar.vue";
import useLayoutStore from "@/store/modules/layout";
import useNamespacesStore from "@/store/modules/namespaces";
import usePrivateKeysStore from "@/store/modules/private_keys";
import useSpinnerStore from "@/store/modules/spinner";

defineOptions({
  inheritAttrs: false,
});

type RouteMeta = {
  title: string;
  icon: string;
  showInSidebar: boolean;
  sidebarOrder: number;
  isPremium?: boolean;
  isBeta?: boolean;
  isHidden?: () => boolean;
};

const router = useRouter();
const layoutStore = useLayoutStore();
const namespacesStore = useNamespacesStore();
const spinnerStore = useSpinnerStore();
const { getPrivateKeyList } = usePrivateKeysStore();
const currentRoute = computed(() => router.currentRoute);
const hasNamespaces = computed(() => namespacesStore.hasNamespaces);
const theme = computed(() => layoutStore.theme);

const { lgAndUp } = useDisplay();

const permanent = computed(() => lgAndUp.value);
const showNavigationDrawer = ref(lgAndUp.value);

const hasSpinner = computed({
  get() { return spinnerStore.status; },
  set(newStatus) { spinnerStore.status = newStatus; },
});

const disableItem = (item: string) => !hasNamespaces.value && item !== "Settings" && item !== "Home";
const isItemHidden = (meta?: RouteMeta) => {
  if (!meta?.isHidden) return false;
  return meta.isHidden();
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
      hidden: isItemHidden(route.meta as RouteMeta),
      sidebarOrder: route.meta?.sidebarOrder as number,
      children: route.children
        ?.filter((child) => child.meta?.showInSidebar)
        .map((child) => ({
          title: child.meta?.title as string,
          path: child.path.startsWith("/") ? child.path : `${route.path}/${child.path}`,
          hidden: isItemHidden(child.meta as RouteMeta),
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

function getFilteredChildren(children: Array<{ title: string; path: string; hidden?: boolean }>) {
  return children.filter((child) => !child.hidden);
}

const visibleItems = computed(() => items.value.filter((item) => !item.hidden));

onMounted(() => { getPrivateKeyList(); });

defineExpose({
  items,
  lgAndUp,
});
</script>

<style lang="scss" scoped>
.container {
  position: relative;
  min-height: calc(100vh - var(--v-layout-top));
  background-image: linear-gradient(180deg, rgb(var(--v-theme-primary), 0.15) 0%, transparent 40%);

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image:
      linear-gradient(to right, rgba(var(--v-border-color), 0.025) 1px, transparent 1px),
      linear-gradient(to bottom, rgba(var(--v-border-color), 0.025) 1px, transparent 1px);
    background-size: 40px 40px, 40px 40px;
    background-position: 0 0, 0 0;
    background-repeat: repeat, repeat;
    mask-image: linear-gradient(to bottom, black 0%, transparent 40%);
    pointer-events: none;
    z-index: 0;
  }

  > * {
    position: relative;
    z-index: 1;
  }
}

.container-light-bg {
  background-image: none;
}

</style>
