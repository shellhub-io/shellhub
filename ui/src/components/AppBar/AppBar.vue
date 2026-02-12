<template>
  <PaywallChat v-model="chatSupportPaywall" />
  <AppBarContent
    show-menu-toggle
    show-support
    @toggle-menu="showNavigationDrawer = !showNavigationDrawer"
    @support-click="openShellhubHelp()"
  >
    <template #left>
      <Namespace />

      <v-breadcrumbs
        :items="breadcrumbItems"
        class="pa-0 mx-4 hidden-xs"
        data-test="breadcrumbs"
      >
        <template #prepend>
          <v-icon
            v-if="breadcrumbItems[0]?.icon"
            :icon="breadcrumbItems[0].icon"
            data-test="breadcrumb-icon"
            size="small"
            class="mr-2"
          />
        </template>
        <template #divider>
          <v-icon
            icon="mdi-chevron-right"
            size="small"
          />
        </template>
      </v-breadcrumbs>
    </template>

    <template #right>
      <DevicesDropdown
        v-if="hasNamespaces"
        v-model="showDevicesDrawer"
        @update:model-value="showInvitationsDrawer = false"
      />

      <InvitationsMenu
        v-if="isCloud"
        v-model="showInvitationsDrawer"
        @update:model-value="showDevicesDrawer = false"
      />

      <UserMenu
        :user-email="userEmail"
        :display-name="currentUser"
        :menu-items="menu"
        :is-dark-mode="isDarkMode"
        @select="handleUserMenuSelect"
        @toggle-dark-mode="toggleDarkMode"
      />
    </template>
  </AppBarContent>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import { useRouter, useRoute, RouteLocationRaw, RouteLocation } from "vue-router";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import handleError from "@/utils/handleError";
import AppBarContent from "@/components/AppBar/AppBarContent.vue";
import UserMenu from "@/components/AppBar/UserMenu.vue";
import DevicesDropdown from "./DevicesDropdown.vue";
import InvitationsMenu from "@/components/Invitations/InvitationsMenu.vue";
import PaywallChat from "../User/PaywallChat.vue";
import Namespace from "@/components/Namespace/Namespace.vue";
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
const { isCommunity, isCloud, isEnterprise } = envVariables;
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();
const tenant = computed(() => authStore.tenantId);
const userEmail = computed(() => authStore.email);
const userId = computed(() => authStore.id);
const currentUser = computed(() => authStore.username);
const isBillingActive = computed(() => billingStore.isActive);
const theme = computed(() => layoutStore.theme);
const hasNamespaces = computed(() => namespacesStore.hasNamespaces);
const isChatCreated = computed(() => supportStore.isChatCreated);
const identifier = computed(() => supportStore.identifier);
const isDarkMode = ref(theme.value === "dark");
const chatSupportPaywall = ref(false);
const showNavigationDrawer = defineModel<boolean>();
const showDevicesDrawer = ref(false);
const showInvitationsDrawer = ref(false);

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

const handleUserMenuSelect = (item: unknown) => {
  void triggerClick(item as MenuItem);
};

const logout = async () => {
  try {
    await router.push({ name: "Login" });
    authStore.logout();
    namespacesStore.namespaceList = [];
    statsStore.stats = {} as IStats;
    if (isChatCreated.value) {
      toggle("close");
      reset();
      supportStore.isChatCreated = false;
    }
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
    case isCloud && isBillingActive.value:
      await openChatwoot();
      break;

    case isCommunity || (isCloud && !isBillingActive.value):
      openPaywall();
      break;

    case isEnterprise:
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

defineExpose({ openShellhubHelp, chatSupportPaywall, breadcrumbItems, showNavigationDrawer });
</script>
