import { RouteRecordRaw, createRouter, createWebHistory } from "vue-router";
import { envVariables } from "../envVariables";
import { store } from "@/store";

const Dashboard = () => import("@/views/Dashboard.vue");
const Devices = () => import("@/views/Devices.vue");
const DeviceList = () => import("@/components/Devices/DeviceList.vue");
const DevicePendingList = () => import("@/components/Devices/DevicePendingList.vue");
const DeviceRejectedList = () => import("@/components/Devices/DeviceRejectedList.vue");
const Containers = () => import("@/views/Containers.vue");
const ContainerList = () => import("@/components/Containers/ContainerList.vue");
const ContainerPendingList = () => import("@/components/Containers/ContainerPendingList.vue");
const ContainerRejectedList = () => import("@/components/Containers/ContainerRejectedList.vue");
const Connectors = () => import("@/views/Connectors.vue");
const detailsConnectors = () => import("@/views/ConnectorDetails.vue");
const DetailsDevice = () => import("@/views/DetailsDevice.vue");
const Sessions = () => import("@/views/Sessions.vue");
const DetailsSessions = () => import("@/views/DetailsSessions.vue");
const FirewallRules = () => import("@/views/FirewallRules.vue");
const PublicKeys = () => import("@/views/PublicKeys.vue");
const Settings = () => import("@/views/Settings.vue");
const SettingProfile = () => import("@/components/Setting/SettingProfile.vue");
const SettingNamespace = () => import("@/components/Setting/SettingNamespace.vue");
const SettingPrivateKeys = () => import("@/components/Setting/SettingPrivateKeys.vue");
const SettingTags = () => import("@/components/Setting/SettingTags.vue");
const SettingBilling = () => import("@/components/Setting/SettingBilling.vue");

export const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "login",
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
      // Check if the user is coming from the login route
      if (from.name === "login") {
        // Allow access to MFA login if the user is coming from the login route
        next();
      } else {
        // Redirect to login if the user is not coming from the login route
        next({ name: "login" });
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
        next({ name: "login" });
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
        next({ name: "login" });
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
    name: "Dashboard",
    component: Dashboard,
  },
  {
    path: "/devices",
    name: "devices",
    component: Devices,
    redirect: {
      name: "listDevices",
    },
    children: [
      {
        path: "",
        name: "listDevices",
        component: DeviceList,
      },
      {
        path: "pending",
        name: "pendingDevices",
        component: DevicePendingList,
      },
      {
        path: "rejected",
        name: "rejectedDevices",
        component: DeviceRejectedList,
      },
    ],
  },
  {
    path: "/containers",
    name: "containers",
    beforeEnter: (to, from, next) => {
      if (!envVariables.isEnterprise && !envVariables.isCloud && envVariables.premiumPaywall) {
        store.commit("users/setShowPaywall", true);
      }
      next();
    },
    component: Containers,
    redirect: {
      name: "listContainers",
    },
    children: [
      {
        path: "",
        name: "listContainers",
        component: ContainerList,
      },
      {
        path: "pending",
        name: "pendingContainers",
        component: ContainerPendingList,
      },
      {
        path: "rejected",
        name: "rejectedContainers",
        component: ContainerRejectedList,
      },
    ],
  },
  {
    path: "/containers/connectors",
    name: "connectors",
    component: Connectors,
    beforeEnter: (to, from, next) => {
      if (!envVariables.isEnterprise && !envVariables.isCloud && envVariables.premiumPaywall) {
        store.commit("users/setShowPaywall", true);
      }
      next();
    },
  },
  {
    path: "/containers/connectors/:id",
    name: "detailsConnectors",
    component: detailsConnectors,
  },
  {
    path: "/device/:id",
    name: "detailsDevice",
    component: DetailsDevice,
  },
  {
    path: "/sessions",
    name: "Sessions",
    component: Sessions,
  },
  {
    path: "/sessions/:id",
    name: "detailsSession",
    component: DetailsSessions,
  },
  {
    path: "/firewall/rules",
    name: "firewalls",
    component: FirewallRules,
    beforeEnter: (to, from, next) => {
      if (!envVariables.isEnterprise && !envVariables.isCloud && envVariables.premiumPaywall) {
        store.commit("users/setShowPaywall", true);
      }
      next();
    },
  },
  {
    path: "/sshkeys/public-keys",
    name: "publicKeys",
    component: PublicKeys,
  },
  {
    path: "/settings",
    name: "settings",
    component: Settings,
    redirect: {
      name: "profileSettings",
    },
    children: [
      {
        path: "profile",
        name: "profileSettings",
        component: SettingProfile,
      },
      {
        path: "namespace-manager",
        name: "namespaceSettings",
        component: SettingNamespace,
      },
      {
        path: "private-keys",
        name: "privateKeysSettings",
        component: SettingPrivateKeys,
      },
      {
        path: "tags",
        name: "tagsSettings",
        component: SettingTags,
      },
      {
        path: "billing",
        name: "billingSettings",
        beforeEnter: (to, from, next) => {
          const enabled = envVariables.billingEnable;
          if (enabled) {
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
    path: "/:catchAll(.*)",
    redirect: { name: "NotFound" },
  },
  {
    path: "/404",
    name: "NotFound",
    component: () => import("../views/NotFound.vue"),
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
    next({ name: "login" }); // Redirect to login page if authentication is required and user is not logged in
  } else if (to.name === "login" && isLoggedIn) {
    next({ path: "/" }); // Redirect from login page to home if user is already logged in
  } else {
    next(); // Continue with the original navigation
  }
});
