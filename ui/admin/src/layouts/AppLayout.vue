<template>
  <v-app-bar
    theme="dark"
    class="bg-v-theme-surface border-b-thin"
    data-test="app-bar"
    flat
    floating
  >
    <v-app-bar-nav-icon
      class="hidden-lg-and-up"
      aria-label="Toggle Menu"
      @click.stop="drawer = !drawer"
    />

    <v-app-bar-title>
      <router-link
        :to="{ name: 'dashboard' }"
        class="text-white text-decoration-none"
      >
        <div class="d-flex">
          <v-img
            :src="Logo"
            max-width="180"
            alt=""
          />
          <span class="mt-4 text-overline">admin</span>
        </div>
      </router-link>
    </v-app-bar-title>

    <v-spacer />

    <v-menu anchor="bottom">
      <template #activator="{ props }">
        <v-chip
          color="primary"
          v-bind="props"
          class="mr-8"
        >
          <v-icon
            left
            class="mr-2"
          >
            mdi-account
          </v-icon>
          {{ currentUser || "ADMIN DF" }}
          <v-icon right>
            mdi-chevron-down
          </v-icon>
        </v-chip>
      </template>
      <v-list class="bg-v-theme-surface">
        <v-list-item
          v-for="item in menu"
          :key="item.title"
          :value="item"
          :data-test="item.title"
          @click="triggerClick(item)"
        >
          <div class="d-flex align-center">
            <div>
              <v-icon :icon="item.icon" />
            </div>

            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </div>
        </v-list-item>

        <v-divider />

        <v-list-item>
          <v-switch
            label="Dark Mode"
            :model-value="isDarkMode"
            data-test="dark-mode-switch"
            color="primary"
            inset
            hide-details
            @change="toggleDarkMode"
          />
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>

  <v-navigation-drawer
    v-if="isLoggedIn"
    v-model="drawer"
    class="bg-v-theme-surface"
    expand-on-hover
  >
    <v-list
      density="compact"
      data-test="list"
    >
      <template
        v-for="item in visibleItems"
        :key="item.title"
      >
        <v-list-group
          v-if="item.children"
          v-model="subMenuState[item.title]"
          prepend-icon="mdi-chevron-down"
          data-test="list-group"
        >
          <template #activator="{ props }">
            <v-list-item
              lines="two"
              v-bind="props"
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
            data-test="list-item"
          >
            <v-list-item-title :data-test="`${child.title}-listItem`">
              {{ child.title }}
            </v-list-item-title>
          </v-list-item>
        </v-list-group>

        <v-list-item
          v-else
          :to="item.path"
          lines="two"
          class="mb-2"
          data-test="list-item"
        >
          <template #prepend>
            <v-icon data-test="icon">
              {{ item.icon }}
            </v-icon>
          </template>
          <v-list-item-title :data-test="`${item.icon}-listItem`">
            {{ item.title }}
          </v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
  </v-navigation-drawer>

  <Snackbar />

  <v-main>
    <slot>
      <v-container
        class="pa-8 container"
        fluid
      >
        <router-view :key="currentRoute.value.path" />
      </v-container>
    </slot>
  </v-main>

  <v-overlay
    v-model="hasSpinner"
    :scrim="false"
    contained
    class="align-center justify-center w-100 h-100"
  >
    <v-progress-circular
      indeterminate
      size="64"
    />
  </v-overlay>
</template>

<script setup lang="ts">
import { watch, ref, computed, reactive } from "vue";
import { RouteLocationRaw, useRouter } from "vue-router";
import useLicenseStore from "@admin/store/modules/license";
import useLayoutStore from "@admin/store/modules/layout";
import useAuthStore from "@admin/store/modules/auth";
import useSpinnerStore from "@/store/modules/spinner";
import Snackbar from "@/components/Snackbar/Snackbar.vue";
import Logo from "../assets/logo-inverted.svg";
import { createNewAdminClient } from "@/api/http";
import { envVariables } from "../envVariables";

type MenuItem = {
  title: string;
  icon: string;
  type: string;
  path: RouteLocationRaw;
  method: () => void;
  hidden?: boolean;
};

type DrawerItem = {
  icon?: string;
  title: string;
  path: RouteLocationRaw;
  hidden?: boolean;
  children?: DrawerItem[];
};

defineOptions({
  inheritAttrs: false,
});

const spinnerStore = useSpinnerStore();
const licenseStore = useLicenseStore();
const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const router = useRouter();
const isLoggedIn = computed(() => authStore.isLoggedIn);

const expiredLicense = computed(() => licenseStore.isExpired);

const hasSpinner = computed(() => spinnerStore.status);
const currentUser = computed(() => authStore.currentUser);
const currentRoute = computed(() => router.currentRoute);
const theme = computed(() => layoutStore.theme);
const isDarkMode = ref(theme.value === "dark");
const drawer = ref(true);

watch(drawer, () => {
  if (window.innerWidth > 1264) {
    drawer.value = true;
  }
});

const logout = async () => {
  authStore.logout();
  await router.push("/login");
  createNewAdminClient();
  layoutStore.layout = "SimpleLayout";
};

const triggerClick = async (item: MenuItem) => {
  switch (item.type) {
    case "path":
      await router.push(item.path);
      break;
    case "method":
      item.method();
      break;
    default:
      break;
  }
};

const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value;
  layoutStore.setTheme(isDarkMode.value ? "dark" : "light");
};

const items = reactive([
  {
    icon: "mdi-view-dashboard",
    title: "Dashboard",
    path: "/",
  },
  {
    icon: "mdi-account",
    title: "Users",
    path: "/users",
  },
  {
    icon: "mdi-developer-board",
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
    path: "/firewall-rules",
  },
  {
    icon: "mdi-login",
    title: "Namespaces",
    path: "/namespaces",
  },
  {
    icon: "mdi-bullhorn",
    title: "Announcements",
    path: "/announcements",
    hidden: !envVariables.announcementsEnable,
  },
  {
    icon: "mdi-cog",
    title: "Settings",
    path: "/settings",
    children: [
      {
        title: "Authentication",
        path: "/settings/authentication",
      },
      {
        title: "License",
        path: "/settings/license",
      },
    ],
  },
]);

const menu = reactive([
  {
    icon: "mdi-license",
    title: "License",
    type: "path",
    path: "/settings/license",
    method: () => {},
  },
  {
    icon: "mdi-logout",
    title: "Logout",
    type: "method",
    path: "",
    method: () => logout,
  },
]);

const subMenuState = reactive<Record<string, boolean>>({});

items.forEach((item) => {
  if (item.children) {
    subMenuState[item.title] = false;
  }
});

const getFilteredChildren = (children: DrawerItem[]) => expiredLicense.value
  ? children.filter((child) => child.title === "License")
  : children.filter((child) => !child.hidden);

const visibleItems = computed(() => {
  if (expiredLicense.value) {
    return items
      .filter((item) => item.title === "Settings")
      .map((item) => ({
        ...item,
        children: item.children?.filter((child) => child.title === "License"),
      }));
  }

  return items.filter((item) => !item.hidden);
});

</script>

<style lang="css" scoped>
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
