import { RouteRecordRaw, createRouter, createWebHistory, RouteLocationNormalized, NavigationGuardNext } from "vue-router";
import { envVariables } from "../envVariables";
import { plugin as snackbar } from "@/plugins/snackbar"; // using direct plugin because inject() doesn't work outside components
import useAuthStore from "@/store/modules/auth";
import useContainersStore from "@/store/modules/containers";
import useDevicesStore from "@/store/modules/devices";
import useLayoutStore, { Layout } from "@/store/modules/layout";
import useNamespacesStore from "@/store/modules/namespaces";
import useUsersStore from "@/store/modules/users";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import useTagsStore from "@/store/modules/tags";
import hasPermission from "@/utils/permission";

export const handleAcceptInvite = async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const namespacesStore = useNamespacesStore();
  try {
    await namespacesStore.lookupUserStatus({
      tenant: (to.query["tenant-id"] || from.query["tenant-id"]) as string,
      id: (to.query["user-id"] || from.query["user-id"]) as string,
      sig: (to.query.sig || from.query.sig) as string,
    });
    const { userStatus } = namespacesStore;
    const isLoggedIn = useAuthStore().isLoggedIn;

    switch (userStatus) {
      case "invited":
        return next({
          path: "/sign-up",
          query: { redirect: to.path, ...to.query },
        });
      case "not-confirmed":
        return next({
          path: "/login",
          query: { redirect: "/accept-invite", ...to.query },
        });
      case "confirmed":
        if (!isLoggedIn) {
          return next({
            path: "/login",
            query: { redirect: "/accept-invite", ...to.query },
          });
        }
        return next();
      default:
        return next();
    }
  } catch {
    snackbar.showError("Failed to accept invitation.");
    next({ name: "Login" });
  }
};

const Home = () => import("@/views/Home.vue");
const Devices = () => import("@/views/Devices.vue");
const DeviceList = () => import("@/components/Devices/DeviceList.vue");
const DevicePendingList = () => import("@/components/Devices/DevicePendingList.vue");
const DeviceRejectedList = () => import("@/components/Devices/DeviceRejectedList.vue");
const Tags = () => import("@/views/Tags.vue");
const Containers = () => import("@/views/Containers.vue");
const ContainerList = () => import("@/components/Containers/ContainerList.vue");
const ContainerPendingList = () => import("@/components/Containers/ContainerPendingList.vue");
const ContainerRejectedList = () => import("@/components/Containers/ContainerRejectedList.vue");
const WebEndpoints = () => import("@/views/WebEndpoints.vue");
const DeviceDetails = () => import("@/views/DetailsDevice.vue");
const Sessions = () => import("@/views/Sessions.vue");
const SessionDetails = () => import("@/views/DetailsSessions.vue");
const FirewallRules = () => import("@/views/FirewallRules.vue");
const PublicKeys = () => import("@/views/PublicKeys.vue");
const PrivateKeys = () => import("@/views/PrivateKeys.vue");
const AcceptInvite = () => import("@/views/NamespaceInviteCard.vue");
const Settings = () => import("@/views/Settings.vue");
const SettingProfile = () => import("@/components/Setting/SettingProfile.vue");
const SettingNamespace = () => import("@/components/Setting/SettingNamespace.vue");
const SettingBilling = () => import("@/components/Setting/SettingBilling.vue");
const TeamMembers = () => import("@/views/TeamMembers.vue");
const TeamApiKeys = () => import("@/views/TeamApiKeys.vue");
const TeamInvitations = () => import("@/views/TeamInvitations.vue");

export const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "Login",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    beforeEnter: async (to, from, next) => {
      await useUsersStore().fetchSystemInfo();
      if (!envVariables.isCloud && !useUsersStore().systemInfo.setup) {
        return next({ name: "Setup" });
      }
      next();
    },
    component: () => import("../views/Login.vue"),
  },
  {
    path: "/system-unavailable",
    name: "System Unavailable",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    beforeEnter: (to, from, next) => {
      // Prevent users from manually navigating to the System Unavailable page
      if (!to.query.redirect) return next({ name: "Home" });
      next();
    },
    component: () => import("../views/UnavailabilityMessage.vue"),
  },
  {
    path: "/mfa-login",
    name: "MfaLogin",
    beforeEnter: (to, from, next) => {
      if (from.name === "Login") {
        next();
      } else {
        next({ name: "Login" });
      }
    },
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/MfaLogin.vue"),
  },
  {
    path: "/recover-mfa",
    name: "RecoverMfa",
    beforeEnter: (to, from, next) => {
      if (from.name === "MfaLogin") {
        next();
      } else {
        next({ name: "Login" });
      }
    },
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../components/AuthMFA/MfaRecover.vue"),
  },
  {
    path: "/recover-mfa/mail-sucessful",
    name: "RecoverMfaMsg",
    beforeEnter: (to, from, next) => {
      if (from.name === "RecoverMfa") {
        next();
      } else {
        next({ name: "Login" });
      }
    },
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../components/AuthMFA/MfaMailRecover.vue"),
  },
  {
    path: "/forgot-pass",
    name: "ForgotPassword",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/ForgotPassword.vue"),
  },
  {
    path: "/validation-account",
    name: "ValidationAccount",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/ValidationAccount.vue"),
  },
  {
    path: "/reset-mfa",
    name: "MfaResetValidation",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/MfaResetValidation.vue"),
  },
  {
    path: "/update-password",
    name: "UpdatePassword",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/UpdatePassword.vue"),
  },
  {
    path: "/sign-up",
    name: "SignUp",
    beforeEnter: async (to, from, next) => {
      await useUsersStore().fetchSystemInfo();
      if (!envVariables.isCloud && !useUsersStore().systemInfo.setup) {
        return next({ name: "Setup" });
      }
      next();
    },
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/SignUp.vue"),
  },
  {
    path: "/setup",
    name: "Setup",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    beforeEnter: async (to, from, next) => {
      await useUsersStore().fetchSystemInfo();
      const forceSetup = to.query.force === "true";
      if (!forceSetup && (envVariables.isCloud || useUsersStore().systemInfo.setup)) {
        next({ name: "Login" });
      } else {
        next();
      }
    },
    component: () => import("../views/Setup.vue"),
  },
  {
    path: "/confirm-account",
    name: "ConfirmAccount",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/ConfirmAccount.vue"),
  },
  {
    path: "/accept-invite",
    name: "AcceptInvite",
    component: AcceptInvite,
    beforeEnter: handleAcceptInvite,
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
  },
  {
    path: "/",
    name: "Home",
    component: Home,
    meta: {
      icon: "mdi-home",
      title: "Home",
      showInSidebar: true,
      sidebarOrder: 1,
    },
  },
  {
    path: "/devices",
    name: "Devices",
    component: Devices,
    beforeEnter: async (to, from, next) => {
      await useDevicesStore().setDeviceListVisibility();
      next();
    },
    redirect: { name: "DeviceList" },
    meta: {
      icon: "mdi-developer-board",
      title: "Devices",
      showInSidebar: true,
      sidebarOrder: 2,
    },
    children: [
      {
        path: "",
        name: "DeviceList",
        component: DeviceList,
        meta: {
          title: "Accepted",
          showInSidebar: false,
        },
      },
      {
        path: "pending",
        name: "DevicePendingList",
        component: DevicePendingList,
        meta: {
          title: "Pending",
          showInSidebar: false,
        },
      },
      {
        path: "rejected",
        name: "DeviceRejectedList",
        component: DeviceRejectedList,
        meta: {
          title: "Rejected",
          showInSidebar: false,
        },
      },
    ],
  },
  {
    path: "/containers",
    name: "Containers",
    component: Containers,
    beforeEnter: async (to, from, next) => {
      await useContainersStore().setContainerListVisibility();
      next();
    },
    redirect: { name: "ContainerList" },
    meta: {
      icon: "mdi-docker",
      title: "Containers",
      showInSidebar: true,
      sidebarOrder: 3,
    },
    children: [
      {
        path: "",
        name: "ContainerList",
        component: ContainerList,
        meta: {
          title: "Accepted",
          showInSidebar: false,
        },
      },
      {
        path: "pending",
        name: "ContainerPendingList",
        component: ContainerPendingList,
        meta: {
          title: "Pending",
          showInSidebar: false,
        },
      },
      {
        path: "rejected",
        name: "ContainerRejectedList",
        component: ContainerRejectedList,
        meta: {
          title: "Rejected",
          showInSidebar: false,
        },
      },
    ],
  },
  {
    path: "/webendpoints",
    name: "WebEndpoints",
    component: WebEndpoints,
    beforeEnter: async (to, from, next) => {
      await useWebEndpointsStore().fetchWebEndpointsList();
      next();
    },
    meta: {
      icon: "mdi-web",
      title: "Web Endpoints",
      showInSidebar: true,
      isBeta: true,
      isHidden: () => !envVariables.hasWebEndpoints,
      sidebarOrder: 4,
    },
  },
  {
    path: "/devices/:identifier",
    name: "DeviceDetails",
    component: DeviceDetails,
    meta: {
      title: "Details",
    },
  },
  {
    path: "/devices/:id/terminal",
    name: "DeviceTerminal",
    component: DeviceDetails,
  },
  {
    path: "/tags",
    name: "Tags",
    component: Tags,
    beforeEnter: async (to, from, next) => {
      await useTagsStore().setTagListVisibility();
      next();
    },
    meta: {
      icon: "mdi-tag-multiple",
      title: "Tags",
    },
  },
  {
    path: "/sessions",
    name: "Sessions",
    component: Sessions,
    meta: {
      icon: "mdi-history",
      title: "Sessions",
      showInSidebar: true,
      sidebarOrder: 6,
    },
  },
  {
    path: "/sessions/:id",
    name: "SessionDetails",
    component: SessionDetails,
    meta: {
      title: "Details",
    },
  },
  {
    path: "/firewall/rules",
    name: "FirewallRules",
    component: FirewallRules,
    beforeEnter: (to, from, next) => {
      if (envVariables.isCommunity && envVariables.premiumPaywall) {
        useUsersStore().showPaywall = true;
      }
      next();
    },
    meta: {
      icon: "mdi-security",
      title: "Firewall Rules",
      showInSidebar: true,
      isPremium: true,
      isHidden: () => envVariables.isCommunity && !envVariables.premiumPaywall,
      sidebarOrder: 7,
    },
  },
  {
    path: "/sshkeys/public-keys",
    name: "PublicKeys",
    component: PublicKeys,
    meta: {
      icon: "mdi-key",
      title: "Public Keys",
      showInSidebar: true,
      sidebarOrder: 8,
    },
  },
  {
    path: "/team",
    name: "Team",
    redirect: { name: "ApiKeys" },
    meta: {
      icon: "mdi-account-group",
      title: "Team",
      showInSidebar: true,
      sidebarOrder: 9,
    },
    children: [
      {
        path: "api-keys",
        name: "ApiKeys",
        component: TeamApiKeys,
        meta: {
          title: "API Keys",
          showInSidebar: true,
        },
      },
      {
        path: "members",
        name: "Members",
        component: TeamMembers,
        meta: {
          title: "Members",
          showInSidebar: true,
        },
      },
      {
        path: "invitations",
        name: "Invitations",
        component: TeamInvitations,
        meta: {
          title: "Invitations",
          showInSidebar: true,
          isHidden: () => !envVariables.isCloud || !hasPermission("namespace:editInvitation"),
        },
      },
    ],
  },
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
    redirect: { name: "SettingProfile" },
    meta: {
      icon: "mdi-cog",
      title: "Settings",
      showInSidebar: true,
      sidebarOrder: 10,
    },
    children: [
      {
        path: "profile",
        name: "SettingProfile",
        component: SettingProfile,
        meta: {
          title: "Profile",
          showInSidebar: true,
        },
      },
      {
        path: "namespace",
        name: "SettingNamespace",
        component: SettingNamespace,
        meta: {
          title: "Namespace",
          showInSidebar: true,
          isHidden: () => localStorage.getItem("tenant") === "",
        },
      },
      {
        path: "private-keys",
        name: "PrivateKeys",
        component: PrivateKeys,
        meta: {
          title: "Private Keys",
          showInSidebar: true,
        },
      },
      {
        path: "billing",
        name: "SettingBilling",
        beforeEnter: (to, from, next) => {
          if (envVariables.isCloud) {
            next();
          } else {
            next("/404");
          }
        },
        component: SettingBilling,
        meta: {
          title: "Billing",
          showInSidebar: true,
          isHidden: () => !(envVariables.isCloud && localStorage.getItem("tenant") !== ""),
        },
      },
    ],
  },
  {
    path: "/404",
    name: "NotFound",
    component: () => import("../views/NotFound.vue"),
  },
  {
    path: "/:catchAll(.*)",
    redirect: { name: "NotFound" },
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(
  async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
    if (to.name !== "System Unavailable") {
      try {
        await useUsersStore().checkHealth();
      } catch {
        return next({
          name: "System Unavailable",
          query: { redirect: to.fullPath },
        });
      }
    }

    const isLoggedIn = useAuthStore().isLoggedIn;
    const requiresAuth = to.meta.requiresAuth ?? true;

    if (!isLoggedIn && requiresAuth) return next({
      name: "Login",
      query: { redirect: to.fullPath },
    });

    const layout = to.meta.layout || "AppLayout";
    useLayoutStore().layout = layout as Layout;

    return next();
  },
);
