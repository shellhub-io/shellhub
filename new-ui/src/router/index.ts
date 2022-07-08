import { RouteRecordRaw, createRouter, createWebHistory } from "vue-router";
import { envVariables } from "../envVariables";


const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "login",
    component: () => import("../views/Login.vue"),
  },
  {
    path: "/forgot-pass",
    name: "ForgotPassword",
    component: () => import("../views/ForgotPassword.vue"),
  },
  {
    path: "/validation-account",
    name: "ValidationAccount",
    component: () => import("../views/ValidationAccount.vue"),
  },
  {
    path: "/update-password",
    name: "UpdatePassword",
    component: () => import("../views/UpdatePassword.vue"),
  },
  {
    path: "/sign-up",
    name: "SignUp",
    component: () => import("../views/SignUp.vue"),
  },
  {
    path: "/",
    name: "Dashboard",
    component: () => import("../views/Dashboard.vue"),
  },
  {
    path: "/devices",
    name: "devices",
    component: () => import("../views/Devices.vue"),
    redirect: {
      name: "listDevices",
    },
    children: [
      {
        path: "",
        name: "listDevices",
        component: () => import("../components/Devices/DeviceList.vue"),
      },
      {
        path: "pending",
        name: "pendingDevices",
        component: () => import("../components/Devices/DevicePendingList.vue"),
      },
      {
        path: "rejected",
        name: "rejectedDevices",
        component: () => import("../components/Devices/DeviceRejectedList.vue"),
      },
    ],
  },
  {
    path: "/device/:id",
    name: "detailsDevice",
    component: () => import("../views/DetailsDevice.vue"),
  },
  {
    path: "/sessions",
    name: "Sessions",
    component: () => import("../views/Sessions.vue"),
  },
  {
    path: "/sessions/:id",
    name: "detailsSession",
    component: () => import("../views/DetailsSessions.vue"),
  },
  {
    path: "/firewall/rules",
    name: "firewalls",
    component: () => import("../views/FirewallRules.vue"),
  },
  {
    path: "/sshkeys/public-keys",
    name: "publicKeys",
    component: () => import("../views/PublicKeys.vue"),
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("../views/Settings.vue"),
    redirect: {
      name: "profileSettings",
    },
    children: [
      {
        path: "profile",
        name: "profileSettings",
        component: () => import("../components/Setting/SettingProfile.vue"),
      },
      {
        path: "namespace-manager",
        name: "namespaceSettings",
        component: () => import("../components/Setting/SettingNamespace.vue"),
      },
      {
        path: "private-keys",
        name: "privateKeysSettings",
        component: () => import("../components/Setting/SettingPrivateKeys.vue"),
      },
      {
        path: "tags",
        name: "tagsSettings",
        component: () => import("../components/Setting/SettingTags.vue"),
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
        component: () => import("../components/Setting/SettingBilling.vue"),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory("/"),
  routes,
});

export default router;
