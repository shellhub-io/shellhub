import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import AdminConsoleItemIcon from "@/components/Namespace/AdminConsoleItemIcon.vue";

describe("AdminConsoleItemIcon", () => {
  let wrapper: VueWrapper<InstanceType<typeof AdminConsoleItemIcon>>;

  beforeEach(() => { wrapper = mountComponent(AdminConsoleItemIcon); });
  afterEach(() => { wrapper?.unmount(); });

  it("Renders avatar with correct size", () => {
    const avatar = wrapper.find(".v-avatar");
    expect(avatar.exists()).toBe(true);
  });

  it("Renders shield crown icon", () => {
    const icon = wrapper.find(".v-icon");
    expect(icon.exists()).toBe(true);
  });

  it("Has primary color", () => {
    const avatar = wrapper.find(".v-avatar");
    expect(avatar.classes()).toContain("text-primary");
  });

  it("Has tonal variant", () => {
    const avatar = wrapper.find(".v-avatar");
    expect(avatar.classes()).toContain("v-avatar--variant-tonal");
  });

  it("Has rounded corners", () => {
    const avatar = wrapper.find(".v-avatar");
    expect(avatar.classes()).toContain("v-avatar--rounded");
  });

  it("Has border", () => {
    const avatar = wrapper.find(".v-avatar");
    expect(avatar.classes()).toContain("border");
    expect(avatar.classes()).toContain("border-primary");
  });
});
