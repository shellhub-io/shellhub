import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import UserStatusChip from "@admin/components/User/UserStatusChip.vue";
import { UserStatus } from "@admin/interfaces/IUser";

describe("UserStatusChip", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserStatusChip>>;

  const mountWrapper = (status: UserStatus) => {
    wrapper = mountComponent(UserStatusChip, { props: { status } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("confirmed status", () => {
    beforeEach(() => mountWrapper("confirmed"));

    it("renders chip with success color", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("color")).toBe("success");
    });

    it("displays confirmed icon", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("prependIcon")).toBe("mdi-checkbox-marked-circle");
    });

    it("displays confirmed label", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.text()).toBe("Confirmed");
    });
  });

  describe("invited status", () => {
    beforeEach(() => mountWrapper("invited"));

    it("renders chip with warning color", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("color")).toBe("warning");
    });

    it("displays invited icon", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("prependIcon")).toBe("mdi-email-alert");
    });

    it("displays invited label", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.text()).toBe("Invited");
    });
  });

  describe("not-confirmed status", () => {
    beforeEach(() => mountWrapper("not-confirmed"));

    it("renders chip with error color", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("color")).toBe("error");
    });

    it("displays not-confirmed icon", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("prependIcon")).toBe("mdi-alert-circle");
    });

    it("displays not-confirmed label", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.text()).toBe("Not Confirmed");
    });
  });

  describe("invalid status fallback", () => {
    beforeEach(() => mountWrapper("invalid-status" as UserStatus));

    it("falls back to not-confirmed color for invalid status", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("color")).toBe("error");
    });

    it("falls back to not-confirmed icon for invalid status", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.props("prependIcon")).toBe("mdi-alert-circle");
    });

    it("falls back to not-confirmed label for invalid status", () => {
      const chip = wrapper.findComponent({ name: "VChip" });
      expect(chip.text()).toBe("Not Confirmed");
    });
  });
});
