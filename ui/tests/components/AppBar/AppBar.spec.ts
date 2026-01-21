import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import { createCleanRouter } from "@tests/utils/router";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import AppBar from "@/components/AppBar/AppBar.vue";
import { Router } from "vue-router";
import useSupportStore from "@/store/modules/support";
import { envVariables } from "@/envVariables";
import { VLayout } from "vuetify/components";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockUser } from "@tests/mocks";

const Component = {
  template: "<v-layout><AppBar v-model=\"showNavigationDrawer\" /></v-layout>",
  data() {
    return { showNavigationDrawer: true };
  },
};

vi.mock("@productdevbook/chatwoot/vue", () => ({
  useChatWoot: () => ({
    setUser: vi.fn(),
    setConversationCustomAttributes: vi.fn(),
    toggle: vi.fn(),
    reset: vi.fn(),
  }),
}));

vi.mock("@/store/api/devices");
vi.mock("@/store/api/namespaces");
vi.mock("@/store/api/stats");
vi.mock("@/store/api/support");

const mockBilling = {
  id: "sub_test",
  active: true,
  status: "active",
  customer_id: "cus_test",
  subscription_id: "sub_test",
  current_period_end: 999999999999,
};

describe("AppBar", () => {
  let wrapper: VueWrapper<unknown>;
  let appBar: VueWrapper<InstanceType<typeof AppBar>>;
  let router: Router;
  let supportStore: ReturnType<typeof useSupportStore>;

  const triggerSupportClick = async () => {
    const appBarContent = appBar.findComponent({ name: "AppBarContent" });
    await appBarContent.vm.$emit("support-click");
    await flushPromises();
  };

  const mountWrapper = async (isCloud = true, hasNamespaces = true) => {
    envVariables.isCloud = isCloud;
    envVariables.isCommunity = !isCloud;
    localStorage.setItem("tenant", "tenant-123");

    router = createCleanRouter();
    await router.push("/");
    await router.isReady();

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        components: { AppBar, "v-layout": VLayout },
      },
      piniaOptions: {
        initialState: {
          auth: mockUser,
          billing: { billing: mockBilling },
          namespaces: {
            namespaceList: hasNamespaces ? [{ name: "test-namespace", tenant_id: "tenant-123" }] : [],
          },
          stats: {
            stats: {
              registered_devices: 10,
              online_devices: 5,
              pending_devices: 2,
              rejected_devices: 1,
              active_sessions: 3,
            },
          },
        },
      },
    });
    appBar = wrapper.findComponent(AppBar);

    supportStore = useSupportStore();

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    localStorage.clear();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the app bar", () => {
      expect(appBar.exists()).toBe(true);
    });

    it("displays the menu toggle button", () => {
      const menuToggle = appBar.find('[data-test="menu-toggle"]');
      expect(menuToggle.exists()).toBe(true);
    });

    it("displays the breadcrumbs", () => {
      const breadcrumbs = appBar.find('[data-test="breadcrumbs"]');
      expect(breadcrumbs.exists()).toBe(true);
    });

    it("displays the support button", () => {
      const supportBtn = appBar.find('[data-test="support-btn"]');
      expect(supportBtn.exists()).toBe(true);
    });

    it("displays the user menu button", () => {
      const userMenuBtn = appBar.find('[data-test="user-menu-btn"]');
      expect(userMenuBtn.exists()).toBe(true);
    });

    it("displays the user icon", () => {
      const userIcon = appBar.find('[data-test="user-icon"]');
      expect(userIcon.exists()).toBe(true);
    });

    it("displays the namespace selector", () => {
      const namespaceSelector = appBar.findComponent({ name: "Namespace" });
      expect(namespaceSelector.exists()).toBe(true);
    });
  });

  describe("conditional rendering - cloud features", () => {
    it("shows devices dropdown when in cloud and has namespaces", async () => {
      await mountWrapper(true, true);
      await flushPromises();

      const devicesDropdown = appBar.findComponent({ name: "DevicesDropdown" });
      expect(devicesDropdown.exists()).toBe(true);
    });

    it("hides devices dropdown when no namespaces exist", async () => {
      await mountWrapper(true, false);
      await flushPromises();

      const devicesDropdown = appBar.findComponent({ name: "DevicesDropdown" });
      expect(devicesDropdown.exists()).toBe(false);
    });

    it("shows invitations menu when in cloud environment", async () => {
      await mountWrapper(true, true);
      const invitationsMenu = appBar.findComponent({ name: "InvitationsMenu" });
      expect(invitationsMenu.exists()).toBe(true);
    });

    it("hides invitations menu when not in cloud environment", async () => {
      await mountWrapper(false, true);
      const invitationsMenu = appBar.findComponent({ name: "InvitationsMenu" });
      expect(invitationsMenu.exists()).toBe(false);
    });
  });

  describe("menu toggle interactions", () => {
    beforeEach(() => mountWrapper());

    it("toggles navigation drawer when menu button is clicked", async () => {
      const initialValue = appBar.vm.showNavigationDrawer;

      const appBarContent = appBar.findComponent({ name: "AppBarContent" });
      await appBarContent.vm.$emit("toggle-menu");
      await flushPromises();

      expect(appBar.vm.showNavigationDrawer).toBe(!initialValue);
    });

    it("toggles drawer state multiple times correctly", async () => {
      const appBarContent = appBar.findComponent({ name: "AppBarContent" });
      const initialState = appBar.vm.showNavigationDrawer;

      await appBarContent.vm.$emit("toggle-menu");
      await flushPromises();
      expect(appBar.vm.showNavigationDrawer).toBe(!initialState);
      await appBarContent.vm.$emit("toggle-menu");
      await flushPromises();
      expect(appBar.vm.showNavigationDrawer).toBe(initialState);
    });
  });

  describe("support button interactions", () => {
    beforeEach(() => mountWrapper(true, true));

    it("calls openShellhubHelp when support button is clicked", async () => {
      const openShellhubHelpSpy = vi.spyOn(appBar.vm, "openShellhubHelp");

      await triggerSupportClick();

      expect(openShellhubHelpSpy).toHaveBeenCalled();
    });

    it("fetches support identifier from store when support is clicked", async () => {
      await triggerSupportClick();
      expect(supportStore.getIdentifier).toHaveBeenCalledWith("tenant-123");
    });
  });

  describe("breadcrumbs", () => {
    beforeEach(async () => {
      await mountWrapper();
      await router.push({ name: "Devices" });
      await flushPromises();
    });

    it("updates breadcrumbs based on current route", () => {
      const breadcrumbs = appBar.find('[data-test="breadcrumbs"]');
      expect(breadcrumbs.exists()).toBe(true);
      // Breadcrumbs should reflect the current route
      expect(appBar.vm.breadcrumbItems.length).toBeGreaterThan(0);
    });

    it("shows icon in breadcrumb when route has icon", () => {
      const icon = appBar.find('[data-test="breadcrumb-icon"]');
      expect(icon.exists()).toBe(true);
    });
  });

  describe("user menu", () => {
    beforeEach(() => mountWrapper());

    it("displays user email in menu", () => {
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      expect(userMenu.props("userEmail")).toBe(mockUser.email);
    });

    it("displays user display name in menu", () => {
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      expect(userMenu.props("displayName")).toBe(mockUser.username);
    });

    it("passes menu items to UserMenu component", () => {
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      expect(userMenu.props("menuItems")).toBeDefined();
      expect(Array.isArray(userMenu.props("menuItems"))).toBe(true);
    });

    it("handles user menu selection", async () => {
      const routerPushSpy = vi.spyOn(router, "push");
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      const mockMenuItem = { title: "Profile", icon: "mdi-account", type: "path", path: { name: "SettingProfile" } };

      await userMenu.vm.$emit("select", mockMenuItem);
      await flushPromises();

      // Should handle the menu item selection (navigate or execute method)
      expect(routerPushSpy).toHaveBeenCalledWith({ name: "SettingProfile" });
    });
  });

  describe("dark mode toggle", () => {
    beforeEach(() => mountWrapper());

    it("passes dark mode state to UserMenu", () => {
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      expect(typeof userMenu.props("isDarkMode")).toBe("boolean");
    });

    it("toggles dark mode when event is emitted", async () => {
      const userMenu = appBar.findComponent({ name: "UserMenu" });
      const initialDarkMode = userMenu.props("isDarkMode");

      await userMenu.vm.$emit("toggle-dark-mode");
      await flushPromises();

      // Dark mode should toggle
      expect(userMenu.props("isDarkMode")).toBe(!initialDarkMode);
    });
  });

  describe("drawer interactions", () => {
    beforeEach(() => mountWrapper(true, true));

    it("closes invitations drawer when devices drawer opens", async () => {
      const devicesDropdown = appBar.findComponent({ name: "DevicesDropdown" });

      await devicesDropdown.vm.$emit("update:modelValue", true);
      await flushPromises();

      const invitationsMenu = appBar.findComponent({ name: "InvitationsMenu" });
      expect(invitationsMenu.props("modelValue")).toBe(false);
    });

    it("closes devices drawer when invitations drawer opens", async () => {
      const invitationsMenu = appBar.findComponent({ name: "InvitationsMenu" });

      await invitationsMenu.vm.$emit("update:modelValue", true);
      await flushPromises();

      const devicesDropdown = appBar.findComponent({ name: "DevicesDropdown" });
      expect(devicesDropdown.props("modelValue")).toBe(false);
    });
  });

  describe("community vs cloud behavior", () => {
    it("shows paywall for community instance", async () => {
      await mountWrapper(false, true);
      vi.mocked(supportStore.getIdentifier).mockRejectedValueOnce(new Error("No identifier"));

      await triggerSupportClick();

      expect(appBar.vm.chatSupportPaywall).toBe(true);
    });

    it("does not show paywall for cloud instance with valid identifier", async () => {
      await mountWrapper(true, true);

      appBar.vm.chatSupportPaywall = false;
      await triggerSupportClick();

      expect(appBar.vm.chatSupportPaywall).toBe(false);
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when support identifier fetch fails", async () => {
      await mountWrapper(true, true);
      vi.mocked(supportStore.getIdentifier).mockRejectedValueOnce(createAxiosError(500, "Internal server error"));

      await triggerSupportClick();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "Failed to open chat support. Please check your account's billing and try again later.");
    });
  });
});
