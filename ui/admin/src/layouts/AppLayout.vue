<template>
  <v-app :theme="getStatusDarkMode">
    <v-app-bar theme="dark">
      <v-app-bar-nav-icon class="hidden-lg-and-up" @click.stop="drawer = !drawer" aria-label="Toggle Menu" />

      <v-app-bar-title>
        <router-link :to="{ name: 'dashboard' }" class="admin-name text-decoration-none">
          <div class="d-flex">
            <v-img
              :src="Logo"
              max-width="180"
              alt="Shell logo, a cloud with the writing 'Admin' on the right side"
            />
            <span class="mt-4 text-overline">admin</span>
          </div>
        </router-link>
      </v-app-bar-title>

      <v-spacer />

      <v-menu anchor="bottom">
        <template v-slot:activator="{ props }">
          <v-chip dark v-bind="props" class="mr-8">
            <v-icon left class="mr-2"> mdi-account </v-icon>
            {{ currentUser || "ADMIN DF" }}
            <v-icon right> mdi-chevron-down </v-icon>
          </v-chip>
        </template>
        <v-list class="mr-8">
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

          <v-list-item density="compact">
            <v-switch
              label="Dark Mode"
              :model-value="isDarkMode"
              @change="toggleDarkMode"
              density="compact"
              data-test="dark-mode-switch"
              color="primary"
              hide-details
            />
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <v-navigation-drawer v-if="isLoggedIn" theme="dark" v-model="drawer" expand-on-hover>
      <v-list density="compact" data-test="list">
        <template v-for="item in visibleItems" :key="item.title">
          <v-list-group
            v-if="item.children"
            prepend-icon="mdi-chevron-down"
            v-model="subMenuState[item.title]"
            data-test="list-group"
          >
            <template v-slot:activator="{ props }">
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

    <SnackbarComponent />

    <v-main>
      <slot>
        <v-container class="pa-8" fluid>
          <router-view :key="currentRoute.value.path" />
        </v-container>
      </slot>
    </v-main>

    <v-overlay v-model="hasSpinner">
      <div class="full-width-height d-flex justify-center align-center">
        <v-progress-circular
          indeterminate
          size="64"
        />
      </div>
    </v-overlay>
  </v-app>
</template>

<script setup lang="ts">
import { watch, ref, computed, reactive } from "vue";
import { RouteLocationRaw, useRouter } from "vue-router";
import { useStore } from "../store";
import Logo from "../assets/logo-inverted.png";
import { createNewClient } from "../api/http";
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

const store = useStore();
const router = useRouter();
const isLoggedIn = computed(() => store.getters["auth/isLoggedIn"]);

const expiredLicense = computed(() => store.getters["license/isExpired"]);

const hasSpinner = computed(() => store.getters["spinner/status"]);
const currentUser = computed(() => store.getters["auth/currentUser"]);
const currentRoute = computed(() => router.currentRoute);
const getStatusDarkMode = computed(() => store.getters["layout/getStatusDarkMode"]);
const isDarkMode = ref(getStatusDarkMode.value === "dark");
const drawer = ref(true);

watch(drawer, () => {
  if (window.innerWidth > 1264) {
    drawer.value = true;
  }
});

const logout = async () => {
  await store.dispatch("auth/logout");
  await router.push("/login");
  createNewClient();
  store.dispatch("layout/setLayout", "SimpleLayout");
};

const triggerClick = (item: MenuItem): void => {
  switch (item.type) {
    case "path":
      router.push(item.path);
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
  store.dispatch("layout/setStatusDarkMode", isDarkMode.value);
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
    icon: "mdi-devices",
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
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    method: () => {},
  },
  {
    icon: "mdi-logout",
    title: "Logout",
    type: "method",
    path: "",
    method: logout,
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
.admin-name {
  color: #fff;
  text-decoration: none;
}

.full-width-height {
  width: 100vw !important;
  height: 100vh !important;
}
</style>
