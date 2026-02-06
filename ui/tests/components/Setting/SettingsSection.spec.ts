import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import SettingsSection from "@/components/Setting/SettingsSection.vue";

describe("SettingsSection", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsSection>>;

  const mountWrapper = () => {
    wrapper = mountComponent(SettingsSection, {
      attrs: { "data-test": "settings-section-list" },
      slots: { default: "<div data-test=\"settings-section-item\">Item</div>" },
    });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => wrapper?.unmount());

  describe("Rendering", () => {
    it("Renders v-card container", () => {
      const card = wrapper.findComponent({ name: "v-card" });
      expect(card.exists()).toBe(true);
      expect(card.props("variant")).toBe("flat");
    });

    it("Renders v-list with border and rounded", () => {
      const list = wrapper.findComponent({ name: "v-list" });
      expect(list.exists()).toBe(true);
      expect(list.props("border")).toBe(true);
      expect(list.props("rounded")).toBe(true);
    });

    it("Forwards attributes to v-list", () => {
      const list = wrapper.find('[data-test="settings-section-list"]');
      expect(list.exists()).toBe(true);
    });
  });

  describe("Slot rendering", () => {
    it("Renders default slot content", () => {
      const item = wrapper.find('[data-test="settings-section-item"]');
      expect(item.exists()).toBe(true);
      expect(item.text()).toBe("Item");
    });
  });
});
