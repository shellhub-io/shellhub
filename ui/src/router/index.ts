import { RouteRecordRaw, createRouter, createWebHistory } from "vue-router";
import { envVariables } from "../envVariables";
import Dashboard from "@/views/Dashboard.vue";
import Devices from "@/views/Devices.vue";
import DeviceList from "@/components/Devices/DeviceList.vue";
import DevicePendingList from "@/components/Devices/DevicePendingList.vue";
import DeviceRejectedList from "@/components/Devices/DeviceRejectedList.vue";
import DetailsDevice from "@/views/DetailsDevice.vue";
import Sessions from "@/views/Sessions.vue";
import DetailsSessions from "@/views/DetailsSessions.vue";
import FirewallRules from "@/views/FirewallRules.vue";
import PublicKeys from "@/views/PublicKeys.vue";
import Settings from "@/views/Settings.vue";
import SettingProfile from "@/components/Setting/SettingProfile.vue";
import SettingNamespace from "@/components/Setting/SettingNamespace.vue";
import SettingPrivateKeys from "@/components/Setting/SettingPrivateKeys.vue";
import SettingTags from "@/components/Setting/SettingTags.vue";
import SettingBilling from "@/components/Setting/SettingBilling.vue";
import { store } from "@/store";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/login",
    name: "login",
    meta: {
      external: true,
    },
    component: () => import("../views/Login.vue"),
  },
  {
    path: "/forgot-pass",
    name: "ForgotPassword",
    meta: {
      external: true,
    },
    component: () => import("../views/ForgotPassword.vue"),
  },
  {
    path: "/validation-account",
    name: "ValidationAccount",
    meta: {
      external: true,
    },
    component: () => import("../views/ValidationAccount.vue"),
  },
  {
    path: "/update-password",
    name: "UpdatePassword",
    meta: {
      external: true,
    },
    component: () => import("../views/UpdatePassword.vue"),
  },
  {
    path: "/sign-up",
    name: "SignUp",
    meta: {
      external: true,
    },
    component: () => import("../views/SignUp.vue"),
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
  history: createWebHistory("/"),
  routes,
});

router.beforeEach((to, _, next) => {
  const isLoggedIn = store.getters["auth/isLoggedIn"];
  const isExternal = to.meta.external;

  if (isExternal) {
    store
      .dispatch("layout/setLayout", "simpleLayout")
      .then(() => {
        next();
      })
      .catch((error) => {
        console.log(error);
        next({ name: "login" });
      });
  } else {
    store
      .dispatch("layout/setLayout", "appLayout")
      .then(() => {
        if (isLoggedIn) {
          next();
        } else {
          next({ name: "login" });
        }
      })
      .catch((error) => {
        console.log(error);
        next({ name: "login" });
      });
  }
});

export default router;
