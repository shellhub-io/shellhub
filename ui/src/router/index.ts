/* eslint-disable camelcase */
import { RouteRecordRaw, createRouter, createWebHistory, RouteLocationNormalized, NavigationGuardNext } from "vue-router";
import { envVariables } from "../envVariables";
import { store } from "@/store";
import { plugin as snackbar } from "@/plugins/snackbar"; // using direct plugin because inject() doesn't work outside components
import useAuthStore from "@/store/modules/auth";

export const handleAcceptInvite = async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  try {
    await store.dispatch("namespaces/lookupUserStatus", {
      tenant: to.query["tenant-id"] || from.query["tenant-id"],
      id: to.query["user-id"] || from.query["user-id"],
      sig: to.query.sig || from.query.sig,
    });
    const userStatus = store.getters["namespaces/getUserStatus"];
    const { isLoggedIn } = useAuthStore();

    switch (userStatus) {
      case "invited":
        next({
          path: "/sign-up",
          query: { redirect: to.path, ...to.query },
        });
        return;
      case "not-confirmed":
        next({
          path: "/login",
          query: { redirect: "/accept-invite", ...to.query },
        });
        return;
      case "confirmed":
        if (!isLoggedIn) {
          next({
            path: "/login",
            query: { redirect: "/accept-invite", ...to.query },
          });
          return;
        }
        next();
        break;
      default:
        break;
    }
    next();
  } catch (error) {
    snackbar.showError("Failed to accept invitation.");
    next({ name: "Login" });
  }
};

const Home = () => import("@/views/Home.vue");
const Devices = () => import("@/views/Devices.vue");
const DeviceList = () => import("@/components/Devices/DeviceList.vue");
const DevicePendingList = () => import("@/components/Devices/DevicePendingList.vue");
const DeviceRejectedList = () => import("@/components/Devices/DeviceRejectedList.vue");
const Containers = () => import("@/views/Containers.vue");
const ContainerList = () => import("@/components/Containers/ContainerList.vue");
const ContainerPendingList = () => import("@/components/Containers/ContainerPendingList.vue");
const ContainerRejectedList = () => import("@/components/Containers/ContainerRejectedList.vue");
const Connectors = () => import("@/views/Connectors.vue");
const WebEndpoints = () => import("@/views/WebEndpoints.vue");
const ConnectorDetails = () => import("@/views/ConnectorDetails.vue");
const DeviceDetails = () => import("@/views/DetailsDevice.vue");
const Sessions = () => import("@/views/Sessions.vue");
const SessionDetails = () => import("@/views/DetailsSessions.vue");
const FirewallRules = () => import("@/views/FirewallRules.vue");
const PublicKeys = () => import("@/views/PublicKeys.vue");
const AcceptInvite = () => import("@/views/NamespaceInviteCard.vue");
const Settings = () => import("@/views/Settings.vue");
const SettingProfile = () => import("@/components/Setting/SettingProfile.vue");
const SettingNamespace = () => import("@/components/Setting/SettingNamespace.vue");
const SettingPrivateKeys = () => import("@/components/Setting/SettingPrivateKeys.vue");
const SettingTags = () => import("@/components/Setting/SettingTags.vue");
const SettingBilling = () => import("@/components/Setting/SettingBilling.vue");
const TeamMembers = () => import("@/views/TeamMembers.vue");
const TeamApiKeys = () => import("@/views/TeamApiKeys.vue");

export const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "Login",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    beforeEnter: (to, from, next) => {
      if (envVariables.isCommunity && !store.getters["users/getSystemInfo"].setup) {
        next({ name: "Setup" });
      }
      next();
    },
    component: () => import("../views/Login.vue"),
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
    beforeEnter: (to, from, next) => {
      if (envVariables.isCommunity && !store.getters["users/getSystemInfo"].setup) {
        next({ name: "Setup" });
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
    beforeEnter: (to, from, next) => {
      if (!envVariables.isCommunity || store.getters["users/getSystemInfo"].setup) {
        next({ name: "Login" });
      }
      next();
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
  },
  {
    path: "/devices",
    name: "Devices",
    component: Devices,
    beforeEnter: async (to, from, next) => {
      await store.dispatch("devices/fetch", {
        page: store.getters["devices/getPage"],
        perPage: store.getters["devices/getPerPage"],
        filter: store.getters["devices/getFilter"],
        status: "",
        committable: false,
      });
      next();
    },
    redirect: { name: "DeviceList" },
    children: [
      {
        path: "",
        name: "DeviceList",
        component: DeviceList,
      },
      {
        path: "pending",
        name: "DevicePendingList",
        component: DevicePendingList,
      },
      {
        path: "rejected",
        name: "DeviceRejectedList",
        component: DeviceRejectedList,
      },
    ],
  },
  {
    path: "/containers",
    name: "Containers",
    component: Containers,
    beforeEnter: async (to, from, next) => {
      await store.dispatch("container/fetch", {
        page: store.getters["container/getPage"],
        perPage: store.getters["container/getPerPage"],
        filter: store.getters["container/getFilter"],
        status: "",
        committable: false,
      });
      next();
    },
    redirect: { name: "ContainerList" },
    children: [
      {
        path: "",
        name: "ContainerList",
        component: ContainerList,
      },
      {
        path: "pending",
        name: "ContainerPendingList",
        component: ContainerPendingList,
      },
      {
        path: "rejected",
        name: "ContainerRejectedList",
        component: ContainerRejectedList,
      },
    ],
  },
  {
    path: "/webendpoints",
    name: "WebEndpoints",
    component: WebEndpoints,
    beforeEnter: async (to, from, next) => {
      await store.dispatch("webEndpoints/get", {
        page: store.getters["webEndpoints/getPage"],
        perPage: store.getters["webEndpoints/getPerPage"],
        filter: store.getters["webEndpoints/getFilter"],
        sortBy: store.getters["webEndpoints/getSortBy"],
        orderBy: store.getters["webEndpoints/getOrderBy"],
      });
      next();
    },
  },
  {
    path: "/connectors",
    name: "Connectors",
    component: Connectors,
    beforeEnter: (to, from, next) => {
      if (envVariables.isCommunity && envVariables.premiumPaywall) {
        store.commit("users/setShowPaywall", true);
      }
      next();
    },
  },
  {
    path: "/connectors/:id",
    name: "ConnectorDetails",
    component: ConnectorDetails,
  },
  {
    path: "/devices/:identifier",
    name: "DeviceDetails",
    component: DeviceDetails,
  },
  {
    path: "/devices/:id/terminal",
    name: "DeviceTerminal",
    component: DeviceDetails,
  },
  {
    path: "/sessions",
    name: "Sessions",
    component: Sessions,
  },
  {
    path: "/sessions/:id",
    name: "SessionDetails",
    component: SessionDetails,
  },
  {
    path: "/firewall/rules",
    name: "FirewallRules",
    component: FirewallRules,
    beforeEnter: (to, from, next) => {
      if (envVariables.isCommunity && envVariables.premiumPaywall) {
        store.commit("users/setShowPaywall", true);
      }
      next();
    },
  },
  {
    path: "/sshkeys/public-keys",
    name: "PublicKeys",
    component: PublicKeys,
  },
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
    redirect: { name: "SettingProfile" },
    children: [
      {
        path: "profile",
        name: "SettingProfile",
        component: SettingProfile,
      },
      {
        path: "namespace",
        name: "SettingNamespace",
        component: SettingNamespace,
      },
      {
        path: "private-keys",
        name: "SettingPrivateKeys",
        component: SettingPrivateKeys,
      },
      {
        path: "tags",
        name: "SettingTags",
        component: SettingTags,
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
      },
    ],
  },
  {
    path: "/team",
    name: "Team",
    redirect: { name: "ApiKeys" },
    children: [
      {
        path: "api-keys",
        name: "ApiKeys",
        component: TeamApiKeys,
      },
      {
        path: "members",
        name: "Members",
        component: TeamMembers,
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
    await store.dispatch("users/fetchSystemInfo");
    const { isLoggedIn } = useAuthStore();
    const requiresAuth = to.meta.requiresAuth ?? true;

    const layout = to.meta.layout || "AppLayout";
    await store.dispatch("layout/setLayout", layout);

    if (!isLoggedIn) {
      if (requiresAuth) {
        return next({
          name: "Login",
          query: { redirect: to.fullPath },
        });
      }
    }

    return next();
  },
);
