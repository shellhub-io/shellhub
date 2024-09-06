/* eslint-disable camelcase */
import { RouteRecordRaw, createRouter, createWebHistory } from "vue-router";
import { envVariables } from "../envVariables";
import { store } from "@/store";

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
const Connections = () => import("@/views/Connection.vue");
const ConnectorDetails = () => import("@/views/ConnectorDetails.vue");
const DeviceDetails = () => import("@/views/DetailsDevice.vue");
const Sessions = () => import("@/views/Sessions.vue");
const SessionDetails = () => import("@/views/DetailsSessions.vue");
const FirewallRules = () => import("@/views/FirewallRules.vue");
const PublicKeys = () => import("@/views/PublicKeys.vue");
const Settings = () => import("@/views/Settings.vue");
const SettingProfile = () => import("@/components/Setting/SettingProfile.vue");
const SettingNamespace = () => import("@/components/Setting/SettingNamespace.vue");
const SettingPrivateKeys = () => import("@/components/Setting/SettingPrivateKeys.vue");
const SettingTags = () => import("@/components/Setting/SettingTags.vue");
const SettingBilling = () => import("@/components/Setting/SettingBilling.vue");
const NamespaceInviteDialog = () => import("@/components/Namespace/NamespaceInviteDialog.vue");

export const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "Login",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
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
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import("../views/SignUp.vue"),
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
    path: "/connection/:token",
    name: "Connection",
    component: Connections,
    meta: {
      noGaps: true,
    },
  },
  {
    path: "/connectors/:id",
    name: "ConnectorDetails",
    component: ConnectorDetails,
  },
  {
    path: "/devices/:id",
    name: "DeviceDetails",
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
    path: "/accept-invite",
    name: "AcceptInvite",
    component: NamespaceInviteDialog,
    beforeEnter: (to, from, next) => {
      store.commit("namespaces/setShowNamespaceInvite", true);
      next();
    },
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
          if (envVariables.billingEnable) {
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

router.beforeEach(async (to, from, next) => {
  const isLoggedIn = store.getters["auth/isLoggedIn"];
  const layout = to.meta.layout || "AppLayout";
  const requiresAuth = to.meta.requiresAuth ?? true;

  await store.dispatch("layout/setLayout", layout);

  // Redirect to the appropriate page based on authentication status
  if (requiresAuth && !isLoggedIn) {
    next({
      name: "Login",
      query: { redirect: to.fullPath },
    }); // Redirect to login page if authentication is required and user is not logged in
  } else if (to.name === "Login" && isLoggedIn) {
    next({ path: "/" }); // Redirect from login page to home if user is already logged in
  } else {
    next(); // Continue with the original navigation
  }
});
