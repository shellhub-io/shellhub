import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import NamespaceListItem from "@/components/Namespace/NamespaceListItem.vue";
import { mockNamespace } from "@tests/mocks";

describe("NamespaceListItem", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceListItem>>;

  const mountWrapper = ({ namespace = mockNamespace, active = false, userId = "user-1" } = {}) => {
    wrapper = mountComponent(NamespaceListItem, { props: { namespace, active, userId } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("List item display", () => {
    it("Renders list item with namespace name", () => {
      mountWrapper();

      const listItem = wrapper.findComponent({ name: "VListItem" });
      expect(listItem.exists()).toBe(true);
      expect(listItem.props("title")).toBe(mockNamespace.name);
    });

    it("Shows active state when active prop is true", () => {
      mountWrapper({ active: true });

      const listItem = wrapper.findComponent({ name: "VListItem" });
      expect(listItem.props("active")).toBe(true);
    });

    it("Shows inactive state when active prop is false", () => {
      mountWrapper();

      const listItem = wrapper.findComponent({ name: "VListItem" });
      expect(listItem.props("active")).toBe(false);
    });

    it("Renders NamespaceChip component", () => {
      mountWrapper();

      const chip = wrapper.findComponent({ name: "NamespaceChip" });
      expect(chip.exists()).toBe(true);
      expect(chip.props("name")).toBe(mockNamespace.name);
    });
  });

  describe("User role display", () => {
    it("Shows owner role with crown icon", () => {
      mountWrapper({ userId: "user-1" });

      expect(wrapper.text()).toContain("owner");
      const icons = wrapper.findAll(".v-icon");
      const hasCrownIcon = icons.some((icon) => icon.classes().includes("mdi-crown"));
      expect(hasCrownIcon).toBe(true);
    });

    it("Shows administrator role with shield icon", () => {
      mountWrapper({ userId: "user-2" });

      expect(wrapper.text()).toContain("admin");
      const icons = wrapper.findAll(".v-icon");
      const hasShieldIcon = icons.some((icon) => icon.classes().includes("mdi-shield-account"));
      expect(hasShieldIcon).toBe(true);
    });

    it("Shows operator role with cog icon", () => {
      mountWrapper({ userId: "user-3" });

      expect(wrapper.text()).toContain("operator");
      const icons = wrapper.findAll(".v-icon");
      const hasCogIcon = icons.some((icon) => icon.classes().includes("mdi-account-cog"));
      expect(hasCogIcon).toBe(true);
    });

    it("Shows observer role with eye icon", () => {
      mountWrapper({ userId: "user-4" });

      expect(wrapper.text()).toContain("observer");
      const icons = wrapper.findAll(".v-icon");
      const hasEyeIcon = icons.some((icon) => icon.classes().includes("mdi-eye"));
      expect(hasEyeIcon).toBe(true);
    });

    it("Does not show role when user is not a member", () => {
      mountWrapper({ userId: "non-member-user" });

      const subtitle = wrapper.findComponent({ name: "VListItemSubtitle" });
      expect(subtitle.exists()).toBe(true);
    });
  });

  describe("Namespace type display", () => {
    it("Shows team type with group icon by default", () => {
      const teamNamespace = { ...mockNamespace, type: "team" as const };
      mountWrapper({ namespace: teamNamespace });

      expect(wrapper.text()).toContain("team");
      const icons = wrapper.findAll(".v-icon");
      const hasGroupIcon = icons.some((icon) => icon.classes().includes("mdi-account-group"));
      expect(hasGroupIcon).toBe(true);
    });

    it("Shows personal type with account icon", () => {
      mountWrapper();

      expect(wrapper.text()).toContain("personal");
      const icons = wrapper.findAll(".v-icon");
      const hasAccountIcon = icons.some((icon) => icon.classes().includes("mdi-account"));
      expect(hasAccountIcon).toBe(true);
    });

    it("Shows team type when type is not specified", () => {
      const { type: _type, ...namespaceWithoutType } = mockNamespace;
      // @ts-expect-error Removing type for test
      mountWrapper({ namespace: namespaceWithoutType });

      expect(wrapper.text()).toContain("team");
    });
  });

  describe("Click handling", () => {
    it("Emits select event when inactive item is clicked", async () => {
      mountWrapper();

      const listItem = wrapper.findComponent({ name: "VListItem" });
      await listItem.trigger("click");

      expect(wrapper.emitted("select")).toBeTruthy();
      expect(wrapper.emitted("select")?.[0]).toEqual([mockNamespace.tenant_id]);
    });

    it("Does not emit select event when active item is clicked", async () => {
      mountWrapper({ active: true });

      const listItem = wrapper.findComponent({ name: "VListItem" });
      await listItem.trigger("click");

      expect(wrapper.emitted("select")).toBeFalsy();
    });
  });
});
