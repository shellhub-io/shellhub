import { createRouter, createWebHistory, NavigationGuardNext, RouteLocationNormalized } from "vue-router";
import Login from "@admin/views/Login.vue";
import Dashboard from "@admin/views/Dashboard.vue";
import Users from "@admin/views/Users.vue";
import SettingsLicense from "@admin/components/Settings/SettingsLicense.vue";
import SettingsAuthentication from "@admin/components/Settings/SettingsAuthentication.vue";
import Namespaces from "@admin/views/Namespaces.vue";
import Settings from "@admin/views/Settings.vue";

import { INotificationsError } from "@admin/interfaces/INotifications";
import { computed } from "vue";
import { store } from "../store";

const routes = [
  {
    path: "/login",
    name: "login",
    component: Login,
    meta: {
      layout: "SimpleLayout",
      requiresAuth: false,
    },
  },
  {
    path: "/",
    name: "dashboard",
    component: Dashboard,
  },
  {
    path: "/users",
    name: "users",
    component: Users,
  },
  {
    path: "/user/:id",
    name: "userDetails",
    component: () => import("@admin/views/UserDetails.vue"),
  },
  {
    path: "/devices",
    name: "devices",
    component: () => import("@admin/views/Device.vue"),
    redirect: {
      name: "listDevices",
    },
    children: [
      {
        path: "",
        name: "listDevices",
        component: () => import("@admin/views/Device.vue"),
      },
    ],
  },
  {
    path: "/device/:id",
    name: "deviceDetails",
    component: () => import("@admin/views/DeviceDetails.vue"),
  },
  {
    path: "/sessions",
    name: "sessions",
    component: () => import("@admin/views/Sessions.vue"),
  },
  {
    path: "/session/:id",
    name: "sessionDetails",
    component: () => import("@admin/views/SessionDetails.vue"),
  },
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
    redirect: { name: "SettingProfile" },
    children: [
      {
        path: "authentication",
        name: "SettingAuthentication",
        component: SettingsAuthentication,
      },
      {
        path: "license",
        name: "SettingLicense",
        component: SettingsLicense,
      },
    ],
  },
  {
    path: "/firewall-rules",
    name: "firewall-rules",
    component: () => import("@admin/views/FirewallRules.vue"),
  },
  {
    path: "/firewall-rules/:id",
    name: "firewallRulesDetails",
    component: () => import("@admin/views/FirewallRulesDetails.vue"),
  },
  {
    path: "/namespaces",
    name: "namespaces",
    component: Namespaces,
  },
  {
    path: "/namespace/:id",
    name: "namespaceDetails",
    component: () => import("@admin/views/NamespaceDetails.vue"),
  },
  {
    path: "/announcements",
    name: "announcements",
    component: () => import("@admin/views/Announcements.vue"),
  },
  {
    path: "/announcement/:uuid",
    name: "announcementDetails",
    component: () => import("@admin/views/AnnouncementDetails.vue"),
  },
  {
    path: "/new-announcement",
    name: "new-announcement",
    component: () => import("@admin/views/NewAnnouncement.vue"),
  },
];

const router = createRouter({
  history: createWebHistory("/admin/"),
  routes,
});

router.beforeEach(
  async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
    const isLoggedIn: boolean = store.getters["auth/isLoggedIn"];
    const requiresAuth = to.meta.requiresAuth ?? true;

    const layout = to.meta.layout || "AppLayout";
    await store.dispatch("layout/setLayout", layout);

    if (!isLoggedIn && requiresAuth) {
      return next({
        name: "login",
        query: { redirect: to.fullPath },
      });
    }

    if (isLoggedIn && !to.meta.requiresAuth) {
      const license = computed(() => store.getters["license/license"]);

      try {
        await store.dispatch("license/get");

        if (license.value.expired && to.name !== "SettingLicense") {
          store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.license);
          return next({ name: "SettingLicense" });
        }
      } catch {
        if (to.name !== "SettingLicense") {
          store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.license);
          return next({ name: "SettingLicense" });
        }
      }
    }

    return next();
  },
);

export default router;
