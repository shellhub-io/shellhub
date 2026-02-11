import { describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import { mountComponent } from "@tests/utils/mount";

describe("NoItemsMessage", () => {
  let wrapper: VueWrapper<InstanceType<typeof NoItemsMessage>>;

  const mountWrapper = (props = {}, slots = {}) => {
    wrapper = mountComponent(NoItemsMessage, {
      props: {
        icon: "mdi-alert-circle",
        item: "devices",
        ...props,
      },
      slots,
    });
  };

  describe("Component rendering", () => {
    it("renders icon", () => {
      mountWrapper();

      const icon = wrapper.find('[data-test="message-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-alert-circle");
    });

    it("renders title with item name", () => {
      mountWrapper();

      const title = wrapper.find('[data-test="message-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Looks like you don't have any devices");
    });

    it("renders skeleton loader background", () => {
      mountWrapper();

      const skeleton = wrapper.find(".v-skeleton-loader");
      expect(skeleton.exists()).toBe(true);
    });

    it("renders content slot when provided", () => {
      mountWrapper({}, {
        content: "<p class=\"test-content\">Add your first device to get started</p>",
      });

      const content = wrapper.find('[data-test="message-content"]');
      expect(content.exists()).toBe(true);
      expect(wrapper.html()).toContain("Add your first device to get started");
    });

    it("renders action slot when provided", () => {
      mountWrapper({}, {
        action: "<button class=\"test-action\">Create Device</button>",
      });

      const actions = wrapper.find(".card-actions");
      expect(actions.exists()).toBe(true);
      expect(wrapper.html()).toContain("Create Device");
    });
  });

  describe("Different prop values", () => {
    it.each([
      { icon: "mdi-key", item: "API keys" },
      { icon: "mdi-account", item: "users" },
      { icon: "mdi-shield", item: "firewall rules" },
      { icon: "mdi-desktop-classic", item: "sessions" },
    ])("renders with different items: $item", ({ icon, item }) => {
      mountWrapper({ icon, item });

      expect(wrapper.find('[data-test="message-icon"]').classes()).toContain(icon);
      expect(wrapper.find('[data-test="message-title"]').text()).toBe(`Looks like you don't have any ${item}`);
    });
  });

  describe("Slot combinations", () => {
    it("renders both content and action slots together", () => {
      mountWrapper({}, {
        content: "<div class=\"content-test\">Content here</div>",
        action: "<div class=\"action-test\">Action here</div>",
      });

      expect(wrapper.html()).toContain("Content here");
      expect(wrapper.html()).toContain("Action here");
    });

    it("renders without any slots", () => {
      mountWrapper();

      const content = wrapper.find('[data-test="message-content"]');
      const actions = wrapper.find(".card-actions");

      expect(content.exists()).toBe(true);
      expect(actions.exists()).toBe(true);
      expect(content.text()).toBe("");
    });
  });
});
