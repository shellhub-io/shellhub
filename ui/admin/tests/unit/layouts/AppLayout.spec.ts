import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { Router } from "vue-router";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import useAuthStore from "@admin/store/modules/auth";
import useLayoutStore from "@/store/modules/layout";
import AppLayout from "@admin/layouts/AppLayout.vue";
import { VApp } from "vuetify/components";

const Component = { template: "<v-app><AppLayout /></v-app>" };

// Mock window.location for router tests
Object.defineProperty(window, "location", {
  value: {
    href: "http://localhost:3000/admin/",
    pathname: "/admin/",
    search: "",
    hash: "",
  },
  writable: true,
});

describe("AppLayout", () => {
  let wrapper: VueWrapper;
  let router: Router;
  let authStore: ReturnType<typeof useAuthStore>;
  let layoutStore: ReturnType<typeof useLayoutStore>;

  const mountWrapper = (initialState = {}) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        components: { AppLayout, "v-app": VApp },
      },
      piniaOptions: {
        initialState: {
          adminAuth: {
            token: "dummy-token",
            currentUser: "admin@example.com",
            isAdmin: true,
          },
          adminLicense: { license: { expired: false } },
          spinner: { status: false },
          layout: { theme: "dark" },
          ...initialState,
        },
      },
    });

    authStore = useAuthStore();
    layoutStore = useLayoutStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering when logged in", () => {
    beforeEach(() => mountWrapper());

    it("renders the navigation drawer", () => {
      const drawer = wrapper.findComponent({ name: "VNavigationDrawer" });
      expect(drawer.exists()).toBe(true);
    });

    it("displays the logo in the drawer toolbar", () => {
      const toolbar = wrapper.find('[data-test="drawer-toolbar"]');
      expect(toolbar.exists()).toBe(true);
      expect(toolbar.findComponent({ name: "VImg" }).exists()).toBe(true);
    });

    it("renders the navigation list", () => {
      const list = wrapper.find('[data-test="list"]');
      expect(list.exists()).toBe(true);
    });

    it("displays all menu items when license is not expired", () => {
      const listItems = wrapper.findAll('[data-test="list-item"]');
      // Dashboard, Users, Devices, Sessions, Firewall Rules, Namespaces, Announcements (if enabled)
      expect(listItems.length).toBeGreaterThan(6);
    });

    it("displays settings menu with children", () => {
      const listGroups = wrapper.findAll('[data-test="list-group"]');
      expect(listGroups.length).toBe(1); // Settings has children
    });

    it("displays AppBarContent component", () => {
      const appBarContent = wrapper.findComponent({ name: "AppBarContent" });
      expect(appBarContent.exists()).toBe(true);
    });

    it("passes correct props to AppBarContent", () => {
      const appBarContent = wrapper.findComponent({ name: "AppBarContent" });
      expect(appBarContent.props("showMenuToggle")).toBe(true);
      expect(appBarContent.props("showSupport")).toBe(true);
    });

    it("displays UserMenu component", () => {
      const userMenu = wrapper.findComponent({ name: "UserMenu" });
      expect(userMenu.exists()).toBe(true);
    });

    it("passes user info to UserMenu", () => {
      const userMenu = wrapper.findComponent({ name: "UserMenu" });
      expect(userMenu.props("userEmail")).toBe("admin@example.com");
      expect(userMenu.props("displayName")).toBe("admin@example.com");
    });

    it("displays Namespace component with admin context", () => {
      const namespace = wrapper.findComponent({ name: "Namespace" });
      expect(namespace.exists()).toBe(true);
      expect(namespace.props("isAdminContext")).toBe(true);
    });

    it("renders main content area", () => {
      const main = wrapper.find('[data-test="main"]');
      expect(main.exists()).toBe(true);
    });

    it("renders container for router view", () => {
      const container = wrapper.find('[data-test="container"]');
      expect(container.exists()).toBe(true);
    });
  });

  describe("rendering when not logged in", () => {
    beforeEach(() => {
      mountWrapper({
        adminAuth: {
          token: "",
          currentUser: "",
          isAdmin: false,
        },
      });
    });

    it("does not render navigation drawer when not logged in", () => {
      const drawer = wrapper.findComponent({ name: "VNavigationDrawer" });
      expect(drawer.exists()).toBe(false);
    });

    it("still renders AppBarContent", () => {
      const appBarContent = wrapper.findComponent({ name: "AppBarContent" });
      expect(appBarContent.exists()).toBe(true);
    });

    it("still renders main content area", () => {
      const main = wrapper.find('[data-test="main"]');
      expect(main.exists()).toBe(true);
    });
  });

  describe("expired license behavior", () => {
    beforeEach(() => {
      mountWrapper({
        adminLicense: {
          isExpired: true,
          license: {},
        },
      });
    });

    it("shows only Settings menu item when license is expired", () => {
      const listItems = wrapper.findAll('[data-test="list-item"]');
      // Only Settings submenu items should be visible
      expect(listItems.length).toBeLessThan(3);
    });

    it("shows only License submenu under Settings when license is expired", () => {
      const licenseItem = wrapper.find('[data-test="License-listItem"]');
      expect(licenseItem.exists()).toBe(true);
    });

    it("does not show Authentication submenu when license is expired", () => {
      const authItem = wrapper.find('[data-test="Authentication-listItem"]');
      expect(authItem.exists()).toBe(false);
    });
  });

  describe("spinner overlay", () => {
    beforeEach(() => {
      mountWrapper({
        spinner: {
          status: true,
        },
      });
    });

    it("shows spinner overlay when spinner status is true", () => {
      const overlay = wrapper.find('[data-test="overlay"]');
      expect(overlay.attributes("style")).not.toContain("display: none");
    });

    it("displays progress circular in spinner overlay", () => {
      const progressCircular = wrapper.find('[data-test="progress-circular"]');
      expect(progressCircular.exists()).toBe(true);
    });
  });

  describe("light mode", () => {
    beforeEach(() => mountWrapper({ layout: { theme: "light" } }));

    it("toggles theme when UserMenu emits toggle-dark-mode", async () => {
      const userMenu = wrapper.findComponent({ name: "UserMenu" });

      await userMenu.vm.$emit("toggle-dark-mode");
      await flushPromises();

      expect(layoutStore.setTheme).toHaveBeenCalled();
    });
  });

  describe("user menu interactions", () => {
    beforeEach(() => {
      mountWrapper();
    });

    it("navigates to license page when license menu item is selected", async () => {
      const pushSpy = vi.spyOn(router, "push");
      const userMenu = wrapper.findComponent({ name: "UserMenu" });

      const licenseMenuItem = {
        icon: "mdi-license",
        title: "License",
        type: "path",
        path: "/settings/license",
        method: () => { },
      };

      await userMenu.vm.$emit("select", licenseMenuItem);
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith("/settings/license");
    });

    it("calls logout when logout menu item is selected", async () => {
      const userMenu = wrapper.findComponent({ name: "UserMenu" });

      const logoutMenuItem = {
        icon: "mdi-logout",
        title: "Logout",
        type: "method",
        path: "",
        method: vi.fn(() => {
          authStore.logout();
          window.location.href = "/login";
        }),
      };

      await userMenu.vm.$emit("select", logoutMenuItem);
      await flushPromises();

      expect(logoutMenuItem.method).toHaveBeenCalled();
    });
  });

  describe("support functionality", () => {
    beforeEach(() => {
      mountWrapper();
    });

    it("opens ShellHub help page when support is clicked", async () => {
      const windowOpenSpy = vi.spyOn(window, "open").mockImplementation(() => null);
      const appBarContent = wrapper.findComponent({ name: "AppBarContent" });

      await appBarContent.vm.$emit("support-click");
      await flushPromises();

      expect(windowOpenSpy).toHaveBeenCalledWith(
        "https://github.com/shellhub-io/shellhub/issues/new/choose",
        "_blank",
      );
    });
  });

  describe("menu navigation", () => {
    beforeEach(() => mountWrapper());

    it("navigates to dashboard when logo is clicked", async () => {
      const logoLink = wrapper.find('[data-test="drawer-toolbar"] a');

      await logoLink.trigger("click");
      await flushPromises();

      expect(logoLink.attributes("href")).toContain("/");
    });

    it("displays menu items with correct icons", () => {
      const icons = wrapper.findAll('[data-test="icon"]');
      expect(icons.length).toBeGreaterThan(0);
    });
  });

  describe("Settings menu behavior", () => {
    beforeEach(() => mountWrapper());

    it("shows both Authentication and License in Settings submenu", () => {
      const authItem = wrapper.find('[data-test="Authentication-listItem"]');
      const licenseItem = wrapper.find('[data-test="License-listItem"]');

      expect(authItem.exists()).toBe(true);
      expect(licenseItem.exists()).toBe(true);
    });

    it("disables Settings menu when user is not admin", async () => {
      wrapper.unmount();
      mountWrapper({
        adminAuth: {
          token: "not-admin-token",
          currentUser: "user@example.com",
          isAdmin: false,
        },
      });

      await flushPromises();

      const listGroup = wrapper.findComponent('[data-test="list-group"]');
      // When the user is not an admin the Settings item is the only one in the list group
      const settingsListItem = listGroup.findComponent({ name: "VListItem" });

      expect(settingsListItem.props("disabled")).toBe(true);
    });
  });
});
