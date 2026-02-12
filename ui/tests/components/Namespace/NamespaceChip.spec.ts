import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import NamespaceChip from "@/components/Namespace/NamespaceChip.vue";

describe("NamespaceChip", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceChip>>;

  const mountWrapper = (name?: string) => {
    wrapper = mountComponent(NamespaceChip, { props: { name } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("Chip rendering", () => {
    it("Renders chip component", () => {
      mountWrapper("TestNamespace");

      const chip = wrapper.find(".v-chip");
      expect(chip.exists()).toBe(true);
    });

    it("Has primary color", () => {
      mountWrapper("TestNamespace");

      const chip = wrapper.find(".v-chip");
      expect(chip.classes()).toContain("text-primary");
    });

    it("Has label variant", () => {
      mountWrapper("TestNamespace");

      const chip = wrapper.find(".v-chip");
      expect(chip.classes()).toContain("v-chip--label");
    });

    it("Has uppercase text", () => {
      mountWrapper("TestNamespace");

      const chip = wrapper.find(".v-chip");
      expect(chip.classes()).toContain("text-uppercase");
    });
  });

  describe("First letter display", () => {
    it("Displays first letter of namespace name", () => {
      mountWrapper("TestNamespace");

      expect(wrapper.text()).toBe("T");
    });

    it("Displays empty string when name is empty", () => {
      mountWrapper("");

      expect(wrapper.text()).toBe("");
    });

    it("Displays empty string when name prop is not provided", () => {
      mountWrapper();

      expect(wrapper.text()).toBe("");
    });

    it("Displays first letter when name has spaces", () => {
      mountWrapper("My Namespace");

      expect(wrapper.text()).toBe("M");
    });

    it("Displays first letter when name has special characters", () => {
      mountWrapper("@namespace");

      expect(wrapper.text()).toBe("@");
    });
  });
});
