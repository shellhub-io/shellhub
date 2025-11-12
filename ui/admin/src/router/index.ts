import { createRouter, createWebHistory, NavigationGuardNext, RouteLocationNormalized } from "vue-router";
import Dashboard from "@admin/views/Dashboard.vue";
import Users from "@admin/views/Users.vue";
import SettingsLicense from "@admin/components/Settings/SettingsLicense.vue";
import SettingsAuthentication from "@admin/components/Settings/SettingsAuthentication.vue";
import Namespaces from "@admin/views/Namespaces.vue";
import Settings from "@admin/views/Settings.vue";
import useLicenseStore from "@admin/store/modules/license";
import useLayoutStore, { Layout } from "@admin/store/modules/layout";
import useAuthStore from "@admin/store/modules/auth";
import { plugin as snackbar } from "@/plugins/snackbar"; // using direct plugin because inject() doesn't work outside components

const routes = [
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
    const licenseStore = useLicenseStore();
    const layoutStore = useLayoutStore();
    const authStore = useAuthStore();

    const requiresAuth = to.meta.requiresAuth ?? true;

    layoutStore.layout = to.meta.layout as Layout || "AppLayout";

    if (!authStore.isLoggedIn && requiresAuth) {
      window.location.href = `/login?redirect=${encodeURIComponent(to.fullPath)}`;
      return;
    }

    if (authStore.isLoggedIn && !to.meta.requiresAuth) {
      const { license, getLicense } = licenseStore;

      try {
        await getLicense();

        if (license.expired && to.name !== "SettingLicense") {
          snackbar.showError("Your license has expired. Please update it and try again.");
          return next({ name: "SettingLicense" });
        }
      } catch {
        if (to.name !== "SettingLicense") {
          snackbar.showError("Failed to get your license info. Please check it and try again.");
          return next({ name: "SettingLicense" });
        }
      }
    }

    return next();
  },
);

export default router;
