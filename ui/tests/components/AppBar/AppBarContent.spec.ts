import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import AppBarContent from "@/components/AppBar/AppBarContent.vue";
import { VLayout } from "vuetify/components";

const Component = {
  template: `<v-layout><AppBarContent
    :show-menu-toggle="showMenuToggle"
    :show-support="showSupport"
    @toggle-menu="$emit('toggle-menu')"
    @support-click="$emit('support-click')"
  >
    <template #left>
      <div data-test="left-slot">Left Content</div>
    </template>
    <template #right>
      <div data-test="right-slot">Right Content</div>
    </template>
  </AppBarContent></v-layout>`,
  props: ["showMenuToggle", "showSupport"],
};

describe("AppBarContent", () => {
  let wrapper: VueWrapper;

  const mountWrapper = (props = {}) => {
    wrapper = mountComponent(Component, {
      props: {
        showMenuToggle: true,
        showSupport: true,
        ...props,
      },
      global: { components: { AppBarContent, "v-layout": VLayout } },
    });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the app bar", () => {
      const appBar = wrapper.find('[data-test="app-bar"]');
      expect(appBar.exists()).toBe(true);
    });

    it("displays the menu toggle button when showMenuToggle is true", () => {
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');
      expect(menuToggle.exists()).toBe(true);
    });

    it("displays the support button when showSupport is true", () => {
      const supportBtn = wrapper.find('[data-test="support-btn"]');
      expect(supportBtn.exists()).toBe(true);
      expect(supportBtn.find(".v-icon").classes()).toContain("mdi-help-circle");
    });

    it("renders left slot content", () => {
      const leftSlot = wrapper.find('[data-test="left-slot"]');
      expect(leftSlot.exists()).toBe(true);
      expect(leftSlot.text()).toBe("Left Content");
    });

    it("renders right slot content", () => {
      const rightSlot = wrapper.find('[data-test="right-slot"]');
      expect(rightSlot.exists()).toBe(true);
      expect(rightSlot.text()).toBe("Right Content");
    });
  });

  describe("conditional rendering", () => {
    it("hides menu toggle when showMenuToggle is false", () => {
      mountWrapper({ showMenuToggle: false });
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');
      expect(menuToggle.exists()).toBe(false);
    });

    it("shows menu toggle when showMenuToggle is true", () => {
      mountWrapper({ showMenuToggle: true });
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');
      expect(menuToggle.exists()).toBe(true);
    });

    it("hides support button when showSupport is false", () => {
      mountWrapper({ showSupport: false });
      const supportBtn = wrapper.find('[data-test="support-btn"]');
      expect(supportBtn.exists()).toBe(false);
    });

    it("shows support button when showSupport is true", () => {
      mountWrapper({ showSupport: true });
      const supportBtn = wrapper.find('[data-test="support-btn"]');
      expect(supportBtn.exists()).toBe(true);
    });

    it("hides both controls when both flags are false", () => {
      mountWrapper({ showMenuToggle: false, showSupport: false });
      expect(wrapper.find('[data-test="menu-toggle"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="support-btn"]').exists()).toBe(false);
    });
  });

  describe("user interactions", () => {
    beforeEach(() => mountWrapper());

    it("emits toggle-menu when menu toggle is clicked", async () => {
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');
      await menuToggle.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("toggle-menu")).toBeTruthy();
      expect(wrapper.emitted("toggle-menu")).toHaveLength(1);
    });

    it("emits support-click when support button is clicked", async () => {
      const supportBtn = wrapper.find('[data-test="support-btn"]');
      await supportBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("support-click")).toBeTruthy();
      expect(wrapper.emitted("support-click")).toHaveLength(1);
    });

    it("emits both events when both buttons are clicked", async () => {
      await wrapper.find('[data-test="menu-toggle"]').trigger("click");
      await wrapper.find('[data-test="support-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.emitted("toggle-menu")).toHaveLength(1);
      expect(wrapper.emitted("support-click")).toHaveLength(1);
    });

    it("emits toggle-menu multiple times when clicked repeatedly", async () => {
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');

      await menuToggle.trigger("click");
      await menuToggle.trigger("click");
      await menuToggle.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("toggle-menu")).toHaveLength(3);
    });
  });

  describe("accessibility", () => {
    beforeEach(() => mountWrapper());

    it("has aria-label for menu toggle", () => {
      const menuToggle = wrapper.find('[data-test="menu-toggle"]');
      expect(menuToggle.attributes("aria-label")).toBe("Toggle Menu");
    });

    it("has aria-label for support button", () => {
      const supportBtn = wrapper.find('[data-test="support-btn"]');
      expect(supportBtn.attributes("aria-label")).toBe("community-help-icon");
    });

    it("support button has tooltip", () => {
      const tooltip = wrapper.findComponent({ name: "VTooltip" });
      expect(tooltip.exists()).toBe(true);
    });
  });
});
