import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import AdminConsoleItem from "@/components/Namespace/AdminConsoleItem.vue";

describe("AdminConsoleItem", () => {
  let wrapper: VueWrapper<InstanceType<typeof AdminConsoleItem>>;

  const mountWrapper = (compact = false) => {
    wrapper = mountComponent(AdminConsoleItem, { props: { compact } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("Compact mode", () => {
    it("Renders compact layout when compact prop is true", () => {
      mountWrapper(true);

      expect(wrapper.find('[data-test="admin-console-item"]').exists()).toBe(false);
      expect(wrapper.text()).toContain("Admin Console");
    });
  });

  describe("Full mode", () => {
    beforeEach(() => mountWrapper());

    it("Renders list item when compact prop is false", () => {
      expect(wrapper.find('[data-test="admin-console-item"]').exists()).toBe(true);
    });

    it("Renders list item when compact prop is not provided", () => {
      expect(wrapper.find('[data-test="admin-console-item"]').exists()).toBe(true);
    });

    it("Displays Admin Console title", () => {
      expect(wrapper.find('[data-test="admin-console-item"]').text()).toContain("Admin Console");
    });

    it("Displays Super Admin role", () => {
      expect(wrapper.text()).toContain("Super Admin");
    });

    it("Displays Instance type", () => {
      expect(wrapper.text()).toContain("Instance");
    });

    it("Displays shield crown icon for Super Admin", () => {
      const icons = wrapper.findAll(".v-icon");
      const hasShieldIcon = icons.some((icon) => icon.classes().includes("mdi-shield-crown"));
      expect(hasShieldIcon).toBe(true);
    });

    it("Displays server icon for Instance", () => {
      const icons = wrapper.findAll(".v-icon");
      const hasServerIcon = icons.some((icon) => icon.classes().includes("mdi-server"));
      expect(hasServerIcon).toBe(true);
    });
  });
});
