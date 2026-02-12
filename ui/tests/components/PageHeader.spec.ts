import { describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import PageHeader from "@/components/PageHeader.vue";
import { mountComponent } from "@tests/utils/mount";

describe("PageHeader", () => {
  let wrapper: VueWrapper<InstanceType<typeof PageHeader>>;

  const mountWrapper = (props = {}, slots = {}) => {
    wrapper = mountComponent(PageHeader, {
      props: {
        icon: "mdi-devices",
        title: "Devices",
        ...props,
      },
      slots,
    });
  };

  describe("Component rendering", () => {
    it("renders icon inside avatar", () => {
      mountWrapper();

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(avatar.html()).toContain("mdi-devices");
    });

    it("renders title", () => {
      mountWrapper();

      const title = wrapper.find(".text-h6");
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Devices");
    });

    it("renders description when provided as prop", () => {
      mountWrapper({ description: "Manage your connected devices" });

      const description = wrapper.find(".text-body-2");
      expect(description.exists()).toBe(true);
      expect(description.text()).toBe("Manage your connected devices");
    });

    it("does not render description when not provided", () => {
      mountWrapper();

      const descriptions = wrapper.findAll(".text-body-2");
      expect(descriptions.length).toBe(0);
    });

    it("renders overline when provided", () => {
      mountWrapper({ overline: "Overview" });

      const overline = wrapper.find(".text-overline");
      expect(overline.exists()).toBe(true);
      expect(overline.text()).toBe("Overview");
    });

    it("does not render overline when not provided", () => {
      mountWrapper();

      const overline = wrapper.find(".text-overline");
      expect(overline.exists()).toBe(false);
    });

    it("applies custom icon color", () => {
      mountWrapper({ iconColor: "error" });

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.classes()).toContain("bg-error");
    });

    it("renders description slot when provided", () => {
      mountWrapper({}, {
        description: "<span class=\"custom-desc\">Custom description content</span>",
      });

      expect(wrapper.html()).toContain("Custom description content");
    });

    it("renders actions slot when provided", () => {
      mountWrapper({}, {
        actions: "<button class=\"custom-action\">Action Button</button>",
      });

      expect(wrapper.html()).toContain("Action Button");
    });

    it("renders default slot when provided", () => {
      mountWrapper({}, {
        default: "<div class=\"extra-content\">Extra content below</div>",
      });

      expect(wrapper.html()).toContain("Extra content below");
    });
  });

  describe("Different prop combinations", () => {
    it("renders with all optional props", () => {
      mountWrapper({
        icon: "mdi-account",
        title: "Users",
        description: "Manage team members",
        overline: "Team",
        iconColor: "primary",
      });

      expect(wrapper.html()).toContain("mdi-account");
      expect(wrapper.find(".text-h6").text()).toBe("Users");
      expect(wrapper.find(".text-body-2").text()).toBe("Manage team members");
      expect(wrapper.find(".text-overline").text()).toBe("Team");
      expect(wrapper.find(".v-avatar").classes()).toContain("bg-primary");
    });

    it.each([
      { icon: "mdi-key", title: "API Keys", iconColor: "secondary" },
      { icon: "mdi-shield", title: "Security", iconColor: "success" },
      { icon: "mdi-cog", title: "Settings", iconColor: "warning" },
    ])("renders with different icons and colors: $title", ({ icon, title, iconColor }) => {
      mountWrapper({ icon, title, iconColor });

      expect(wrapper.html()).toContain(icon);
      expect(wrapper.find(".text-h6").text()).toBe(title);
      expect(wrapper.find(".v-avatar").classes()).toContain(`bg-${iconColor}`);
    });
  });

  describe("Slot priorities", () => {
    it("prefers description slot over description prop", () => {
      mountWrapper(
        { description: "Prop description" },
        { description: "<span class=\"slot-desc\">Slot description</span>" },
      );

      expect(wrapper.html()).toContain("Slot description");
      expect(wrapper.html()).toContain("Prop description");
    });

    it("renders all slots together", () => {
      mountWrapper({}, {
        description: "<div>Slot description</div>",
        actions: "<button>Action</button>",
        default: "<div>Default content</div>",
      });

      expect(wrapper.html()).toContain("Slot description");
      expect(wrapper.html()).toContain("Action");
      expect(wrapper.html()).toContain("Default content");
    });
  });

  describe("Responsive behavior", () => {
    it("applies correct responsive classes", () => {
      mountWrapper();

      const container = wrapper.find(".d-flex");
      expect(container.classes()).toContain("flex-column");
      expect(container.classes()).toContain("flex-sm-row");
    });

    it("applies correct responsive text classes to title", () => {
      mountWrapper();

      const title = wrapper.find(".text-h6");
      expect(title.classes()).toContain("text-h6");
      expect(title.classes()).toContain("text-sm-h5");
    });
  });
});
