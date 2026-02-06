import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import UserMenu from "@/components/AppBar/UserMenu.vue";
import { VSwitch } from "vuetify/components";

type MenuItem = {
  title: string;
  icon: string;
  type?: string;
  path?: string;
};

const mockMenuItems: MenuItem[] = [
  { title: "Profile", icon: "mdi-account", type: "path", path: "/settings/profile" },
  { title: "Settings", icon: "mdi-cog", type: "path", path: "/settings" },
  { title: "Logout", icon: "mdi-logout", type: "method" },
];

describe("UserMenu", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserMenu>>;
  let menu: DOMWrapper<Element>;

  const openMenu = async () => {
    const menuButton = wrapper.find('[data-test="user-menu-btn"]');
    await menuButton.trigger("click");
    await flushPromises();
    menu = new DOMWrapper(document.body).find(".v-card");
  };

  const mountWrapper = (props = {}) => {
    wrapper = mountComponent(UserMenu, {
      props: {
        userEmail: "test@example.com",
        displayName: "Test User",
        menuItems: mockMenuItems,
        isDarkMode: false,
        ...props,
      },
    });
  };

  afterEach(() => {
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the user menu button", () => {
      const menuButton = wrapper.find('[data-test="user-menu-btn"]');
      expect(menuButton.exists()).toBe(true);
    });

    it("displays user icon in the button", () => {
      const userIcon = wrapper.find('[data-test="user-icon"]');
      expect(userIcon.exists()).toBe(true);
    });

    it("shows menu when button is clicked", async () => {
      await openMenu();
      expect(menu.exists()).toBe(true);
    });
  });

  describe("menu content", () => {
    beforeEach(async () => {
      mountWrapper();
      await openMenu();
    });

    it("displays large user icon in menu header", () => {
      const largeIcon = menu.find('[data-test="user-icon-large"]');
      expect(largeIcon.exists()).toBe(true);
    });

    it("shows display name as primary label", () => {
      expect(menu.text()).toContain("Test User");
    });

    it("shows email as secondary label when different from display name", () => {
      expect(menu.text()).toContain("test@example.com");
    });

    it("displays all menu items", () => {
      mockMenuItems.forEach((item) => {
        const menuItem = menu.find(`[data-test="${item.title}"]`);
        expect(menuItem.exists()).toBe(true);
        expect(menuItem.text()).toContain(item.title);
      });
    });

    it("shows correct icons for menu items", () => {
      mockMenuItems.forEach((item) => {
        const menuItem = menu.find(`[data-test="${item.title}"]`);
        const icon = menuItem.find(".v-icon");
        expect(icon.classes()).toContain(item.icon);
      });
    });

    it("displays dark mode toggle", () => {
      const darkModeSwitch = menu.find('[data-test="dark-mode-switch"]');
      expect(darkModeSwitch.exists()).toBe(true);
    });

    it("shows Light Mode text when dark mode is off", () => {
      expect(menu.text()).toContain("Light Mode");
    });
  });

  describe("user display variations", () => {
    it("shows only email when no display name is provided", async () => {
      mountWrapper({ displayName: undefined });
      await openMenu();

      expect(menu.text()).toContain("test@example.com");
      // Should not show email twice
      const emailOccurrences = menu.text().split("test@example.com").length - 1;
      expect(emailOccurrences).toBe(1);
    });

    it("shows only display name when same as email", async () => {
      mountWrapper({ userEmail: "Test User", displayName: "Test User" });
      await openMenu();

      expect(menu.text()).toContain("Test User");
      // Should not show twice
      const nameOccurrences = menu.text().split("Test User").length - 1;
      expect(nameOccurrences).toBe(1);
    });

    it("shows User as fallback when no email or display name", async () => {
      mountWrapper({ userEmail: "", displayName: undefined });
      await openMenu();

      expect(menu.text()).toContain("User");
    });
  });

  describe("dark mode toggle", () => {
    it("shows Dark Mode text when dark mode is on", async () => {
      mountWrapper({ isDarkMode: true });
      await openMenu();

      expect(menu.text()).toContain("Dark Mode");
    });

    it("displays switch in off state when dark mode is false", async () => {
      mountWrapper({ isDarkMode: false });
      await openMenu();

      const darkModeSwitch = menu.findComponent('[data-test="dark-mode-switch"]') as VueWrapper<VSwitch>;
      expect(darkModeSwitch.props("modelValue")).toBe(false);
    });

    it("displays switch in on state when dark mode is true", async () => {
      mountWrapper({ isDarkMode: true });
      await openMenu();

      const darkModeSwitch = menu.findComponent('[data-test="dark-mode-switch"]') as VueWrapper<VSwitch>;
      expect(darkModeSwitch.props("modelValue")).toBe(true);
    });

    it("emits toggle-dark-mode when switch item is clicked", async () => {
      mountWrapper();
      await openMenu();

      const darkModeItem = menu.findAll(".v-list-item").at(-1);
      await darkModeItem?.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("toggle-dark-mode")).toBeTruthy();
      expect(wrapper.emitted("toggle-dark-mode")).toHaveLength(1);
    });
  });

  describe("menu item interactions", () => {
    beforeEach(async () => {
      mountWrapper();
      await openMenu();
    });

    it("emits select event when Profile is clicked", async () => {
      const profileItem = menu.find('[data-test="Profile"]');
      await profileItem.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("select")).toBeTruthy();
      expect(wrapper.emitted("select")?.[0]).toEqual([mockMenuItems[0]]);
    });

    it("emits select event when Settings is clicked", async () => {
      const settingsItem = menu.find('[data-test="Settings"]');
      await settingsItem.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("select")).toBeTruthy();
      expect(wrapper.emitted("select")?.[0]).toEqual([mockMenuItems[1]]);
    });

    it("emits select event when Logout is clicked", async () => {
      const logoutItem = menu.find('[data-test="Logout"]');
      await logoutItem.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("select")).toBeTruthy();
      expect(wrapper.emitted("select")?.[0]).toEqual([mockMenuItems[2]]);
    });

    it("emits select event with correct menu item data", async () => {
      const profileItem = menu.find('[data-test="Profile"]');
      await profileItem.trigger("click");
      await flushPromises();

      const emittedItem = wrapper.emitted("select")?.[0]?.[0] as MenuItem;
      expect(emittedItem.title).toBe("Profile");
      expect(emittedItem.icon).toBe("mdi-account");
      expect(emittedItem.type).toBe("path");
      expect(emittedItem.path).toBe("/settings/profile");
    });
  });

  describe("empty menu items", () => {
    beforeEach(async () => {
      mountWrapper({ menuItems: [] });
      await openMenu();
    });

    it("renders without menu items", () => {
      expect(menu.exists()).toBe(true);
      expect(menu.text()).toContain("Test User");
    });

    it("still shows dark mode toggle when no menu items", () => {
      const darkModeSwitch = menu.find('[data-test="dark-mode-switch"]');
      expect(darkModeSwitch.exists()).toBe(true);
    });
  });
});
