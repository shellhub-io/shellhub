import { RouteRecordRaw, createRouter, createWebHistory } from "vue-router";
import { envVariables } from "../envVariables";
import { store } from "@/store";

const Dashboard = () => import(/* webpackChunkName: "dashboard" */ "@/views/Dashboard.vue");
const Devices = () => import(/* webpackChunkName: "devices" */ "@/views/Devices.vue");
const DeviceList = () => import(/* webpackChunkName: "devices" */ "@/components/Devices/DeviceList.vue");
const DevicePendingList = () => import(/* webpackChunkName: "devices" */ "@/components/Devices/DevicePendingList.vue");
const DeviceRejectedList = () => import(/* webpackChunkName: "devices" */ "@/components/Devices/DeviceRejectedList.vue");
const DetailsDevice = () => import(/* webpackChunkName: "device" */ "@/views/DetailsDevice.vue");
const Sessions = () => import(/* webpackChunkName: "sessions" */ "@/views/Sessions.vue");
const DetailsSessions = () => import(/* webpackChunkName: "sessions" */ "@/views/DetailsSessions.vue");
const FirewallRules = () => import(/* webpackChunkName: "firewall-rules" */ "@/views/FirewallRules.vue");
const PublicKeys = () => import(/* webpackChunkName: "public-keys" */ "@/views/PublicKeys.vue");
const Settings = () => import(/* webpackChunkName: "settings" */ "@/views/Settings.vue");
const SettingProfile = () => import(/* webpackChunkName: "settings" */ "@/components/Setting/SettingProfile.vue");
const SettingNamespace = () => import(/* webpackChunkName: "settings" */ "@/components/Setting/SettingNamespace.vue");
const SettingPrivateKeys = () => import(/* webpackChunkName: "settings" */ "@/components/Setting/SettingPrivateKeys.vue");
const SettingTags = () => import(/* webpackChunkName: "settings" */ "@/components/Setting/SettingTags.vue");
const SettingBilling = () => import(/* webpackChunkName: "settings" */ "@/components/Setting/SettingBilling.vue");

const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "login",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import(/* webpackChunkName: "login" */ "../views/Login.vue"),
  },
  {
    path: "/forgot-pass",
    name: "ForgotPassword",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import(/* webpackChunkName: "forgot-password" */ "../views/ForgotPassword.vue"),
  },
  {
    path: "/validation-account",
    name: "ValidationAccount",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import(/* webpackChunkName: "validation-account" */ "../views/ValidationAccount.vue"),
  },
  {
    path: "/update-password",
    name: "UpdatePassword",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import(/* webpackChunkName: "update-password" */ "../views/UpdatePassword.vue"),
  },
  {
    path: "/sign-up",
    name: "SignUp",
    meta: {
      layout: "LoginLayout",
      requiresAuth: false,
    },
    component: () => import(/* webpackChunkName: "sign-up" */ "../views/SignUp.vue"),
  },
  {
    path: "/confirm-account",
    name: "ConfirmAccount",
    meta: {
      layout: "LoginLayout",
    },
    component: () => import(/* webpackChunkName: "confirm-account" */ "../views/ConfirmAccount.vue"),
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
            next("/invalid");
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

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (route) => {
  const isLoggedIn = store.getters["auth/isLoggedIn"];
  // defaults to "AppLayout" if route doesn't requires a custom layout
  const layout = route.meta.layout || "AppLayout";
  const requiresAuth = route.meta.requiresAuth ?? true;

  await store.dispatch("layout/setLayout", layout);

  // redirect to login page if the user was not logged in and auth is required
  if (!isLoggedIn && requiresAuth) {
    return { name: "login" };
  }

  return true;
});

export default router;
