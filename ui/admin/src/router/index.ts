import { createRouter, createWebHistory } from "vue-router";
import Login from "@admin/views/Login.vue";
import Dashboard from "@admin/views/Dashboard.vue";
import Users from "@admin/views/Users.vue";
import SettingsLicense from "@admin/components/Settings/SettingsLicense.vue";
import SettingsAuthentication from "@admin/components/Settings/SettingsAuthentication.vue";
import Namespaces from "@admin/views/Namespaces.vue";
import Settings from "@admin/views/Settings.vue";

import { store } from "../store";

const routes = [
  {
    path: "/login",
    name: "login",
    component: Login,
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

router.beforeEach((to, from, next) => {
  if (to.path !== "/login") {
    if (store.getters["auth/isLoggedIn"]) {
      return next();
    }
    return next(`/login?redirect=${to.path}`);
  }
  if (store.getters["auth/isLoggedIn"]) {
    if (to.path === "/login" && to.query.token) {
      return next();
    }
    return next("/");
  }

  return next();
});

export default router;
