<template>
  <PaywallChat v-model="chatSupportPaywall" />
  <v-app-bar
    flat
    floating
    class="bg-v-theme-surface border-b-thin"
    data-test="app-bar"
  >
    <v-app-bar-nav-icon
      class="hidden-lg-and-up"
      @click.stop="showNavigationDrawer = !showNavigationDrawer"
      aria-label="Toggle Menu"
      data-test="menu-toggle"
    />
    <v-breadcrumbs :items="breadcrumbItems" class="hidden-md-and-down ml-2" data-test="breadcrumbs">
      <template v-slot:prepend>
        <v-icon v-if="breadcrumbItems[0]?.icon" :icon="breadcrumbItems[0].icon" size="small" class="mr-2" />
      </template>
      <template v-slot:divider>
        <v-icon icon="mdi-chevron-right" size="small" />
      </template>
    </v-breadcrumbs>

    <v-spacer />

    <v-tooltip
      location="bottom"
      class="text-center"
    >
      <template v-slot:activator="{ props }">
        <v-btn
          v-bind="props"
          size="medium"
          color="primary"
          aria-label="community-help-icon"
          icon="mdi-help-circle"
          @click="openShellhubHelp()"
          data-test="support-btn"
        />
      </template>
      <span>Need assistance? Click here for support.</span>
    </v-tooltip>

    <NotificationsMenu data-test="notification-component" />

    <v-menu>
      <template v-slot:activator="{ props }">
        <v-btn
          color="primary"
          v-bind="props"
          append-icon="mdi-menu-down"
          class="pl-2 pr-2 mr-4"
          data-test="user-menu-btn"
        >
          <UserIcon size="1.5rem" :email="userEmail" data-test="user-icon" />
        </v-btn>
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
            <v-icon
              :icon="item.icon"
              class="mr-2"
            />

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
            :onchange="toggleDarkMode"
            data-test="dark-mode-switch"
            density="comfortable"
            color="primary"
            inset
            hide-details
          />
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import { useRouter, useRoute, RouteLocationRaw, RouteLocation } from "vue-router";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import handleError from "@/utils/handleError";
import UserIcon from "../User/UserIcon.vue";
import NotificationsMenu from "./Notifications/NotificationsMenu.vue";
import PaywallChat from "../User/PaywallChat.vue";
import { envVariables } from "@/envVariables";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useLayoutStore from "@/store/modules/layout";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";
import { IStats } from "@/interfaces/IStats";
import useSupportStore from "@/store/modules/support";

type MenuItem = {
  title: string;
  icon: string;
  type: string;
  path: RouteLocationRaw;
  method: () => void;
};

type BreadcrumbItem = {
  title: string;
  href: string;
  icon?: string;
};

defineOptions({
  inheritAttrs: false,
});

const { setUser, setConversationCustomAttributes, toggle, reset } = useChatWoot();
const authStore = useAuthStore();
const billingStore = useBillingStore();
const layoutStore = useLayoutStore();
const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const supportStore = useSupportStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();
const tenant = computed(() => authStore.tenantId);
const userEmail = computed(() => authStore.email);
const userId = computed(() => authStore.id);
const currentUser = computed(() => authStore.username);
const isBillingActive = computed(() => billingStore.isActive);
const theme = computed(() => layoutStore.theme);
const isChatCreated = computed(() => supportStore.isChatCreated);
const identifier = computed(() => supportStore.identifier);
const isDarkMode = ref(theme.value === "dark");
const chatSupportPaywall = ref(false);
const showNavigationDrawer = defineModel<boolean>();

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

const logout = async () => {
  try {
    authStore.logout();
    namespacesStore.namespaceList = [];
    statsStore.stats = {} as IStats;
    if (isChatCreated.value) {
      toggle("close");
      reset();
      supportStore.isChatCreated = false;
    }
    await router.push({ name: "Login" });
  } catch (error: unknown) {
    handleError(error);
  }
};

const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value;
  layoutStore.setTheme(isDarkMode.value ? "dark" : "light");
};

const openChatwoot = async (): Promise<void> => {
  try {
    await supportStore.getIdentifier(tenant.value);

    setUser(userId.value, {
      name: currentUser.value,
      email: userEmail.value,
      identifier_hash: identifier.value,
    });

    window.addEventListener(
      "chatwoot:on-message",
      () => {
        setConversationCustomAttributes({
          namespace: namespacesStore.currentNamespace.name,
          tenant: tenant.value,
          domain: window.location.hostname,
        });
      },
      { once: true },
    );

    const holder = document.querySelector(".woot-widget-holder");
    if (holder) {
      window.dispatchEvent(new CustomEvent("chatwoot:ready"));
    }

    supportStore.isChatCreated = true;
    toggle("open");
  } catch (error) {
    snackbar.showError("Failed to open chat support. Please check your account's billing and try again later.");
    handleError(error);
  }
};

const openPaywall = (): void => {
  chatSupportPaywall.value = true;
};

const redirectToGitHub = (): void => {
  window.open("https://github.com/shellhub-io/shellhub/issues/new/choose", "_blank");
};

const openShellhubHelp = async (): Promise<void> => {
  switch (true) {
    case envVariables.isCloud && isBillingActive.value:
      await openChatwoot();
      break;

    case envVariables.isCommunity || (envVariables.isCloud && !isBillingActive.value):
      openPaywall();
      break;

    case envVariables.isEnterprise:
      redirectToGitHub();
      break;

    default:
      snackbar.showError("Your environment configuration is not supported.");
      throw new Error("Unsupported environment configuration.");
  }
};

const menu = [
  {
    title: "Settings",
    type: "path",
    path: "/Settings",
    icon: "mdi-cog",
    // eslint-disable-next-line no-void
    method: () => void 0,
  },
  {
    title: "Logout",
    type: "method",
    icon: "mdi-logout",
    path: "",
    method: logout,
  },
];

const getParentRoute = (path: string) => {
  const routes = router.getRoutes();
  const parentPath = path.substring(0, path.lastIndexOf("/"));
  return routes.find((r) => r.path === parentPath);
};

const generateBreadcrumbs = (route: RouteLocation): BreadcrumbItem[] => {
  const breadcrumbs: BreadcrumbItem[] = [];
  const seenPaths = new Set<string>();

  route.matched.forEach((match) => {
    if (match.name && !seenPaths.has(match.path)) {
      seenPaths.add(match.path);

      const routeName = match.name as string;
      const title = (match.meta?.title as string) || routeName.replace(/([a-z])([A-Z])/g, "$1 $2");
      const icon = match.meta?.icon as string;

      // If this is a child route, add parent first if not already added
      const parentPath = match.path.substring(0, match.path.lastIndexOf("/"));
      if (parentPath && !seenPaths.has(parentPath)) {
        const parent = getParentRoute(match.path);

        if (parent?.meta?.title) {
          seenPaths.add(parentPath);
          breadcrumbs.push({
            title: parent.meta.title as string,
            href: parentPath,
            icon: parent.meta.icon as string,
          });
        }
      }

      breadcrumbs.push({
        title,
        href: match.path,
        icon,
      });
    }
  });

  return breadcrumbs;
};

const breadcrumbItems = computed(() => generateBreadcrumbs(route));

defineExpose({ openShellhubHelp, chatSupportPaywall, logout, isDarkMode, breadcrumbItems, currentUser, identifier });
</script>
