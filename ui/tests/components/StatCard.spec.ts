import { describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import StatCard from "@/components/StatCard.vue";
import { mountComponent } from "@tests/utils/mount";
import { StatCardItem } from "@/interfaces/IStats";

describe("StatCard", () => {
  let wrapper: VueWrapper<InstanceType<typeof StatCard>>;

  const defaultProps: StatCardItem = {
    title: "Active Devices",
    icon: "mdi-devices",
    buttonLabel: "View Devices",
    path: "/devices",
    stat: 42,
  };

  const mountWrapper = (props: Partial<StatCardItem> = {}) => {
    wrapper = mountComponent(StatCard, {
      props: { ...defaultProps, ...props },
    });
  };

  describe("Component rendering", () => {
    it("renders icon inside avatar", () => {
      mountWrapper();

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(wrapper.html()).toContain("mdi-devices");
    });

    it("renders title", () => {
      mountWrapper();

      const title = wrapper.find(".v-card-title");
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Active Devices");
    });

    it("renders stat value", () => {
      mountWrapper();

      const stat = wrapper.find(".v-card-subtitle");
      expect(stat.exists()).toBe(true);
      expect(stat.text()).toBe("42");
    });

    it("renders button with correct label", () => {
      mountWrapper();

      const button = wrapper.findComponent({ name: "VBtn" });
      expect(button.exists()).toBe(true);
      expect(button.text()).toBe("View Devices");
    });

    it("renders button with correct route", () => {
      mountWrapper();

      const button = wrapper.findComponent({ name: "VBtn" });
      expect(button.props("to")).toBe("/devices");
    });
  });

  describe("Different prop values", () => {
    it.each([
      { title: "Online Sessions", icon: "mdi-connection", stat: 15, buttonLabel: "View Sessions", path: "/sessions" },
      { title: "Pending Requests", icon: "mdi-clock", stat: 0, buttonLabel: "View Pending", path: "/pending" },
      { title: "Total Users", icon: "mdi-account-group", stat: 1337, buttonLabel: "Manage Users", path: "/users" },
    ])("renders with different props: $title", ({ title, icon, stat, buttonLabel, path }) => {
      mountWrapper({ title, icon, stat, buttonLabel, path });

      expect(wrapper.html()).toContain(icon);
      expect(wrapper.find(".v-card-title").text()).toBe(title);
      expect(wrapper.find(".v-card-subtitle").text()).toBe(stat.toString());
      const button = wrapper.findComponent({ name: "VBtn" });
      expect(button.text()).toBe(buttonLabel);
      expect(button.props("to")).toBe(path);
    });

    it("handles zero stat value", () => {
      mountWrapper({ stat: 0 });

      const stat = wrapper.find(".v-card-subtitle");
      expect(stat.text()).toBe("0");
    });

    it("handles large stat values", () => {
      mountWrapper({ stat: 999999 });

      const stat = wrapper.find(".v-card-subtitle");
      expect(stat.text()).toBe("999999");
    });
  });
});
