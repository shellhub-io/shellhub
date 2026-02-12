import { VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import { mountComponent } from "@tests/utils/mount";
import { BasicRole } from "@/interfaces/INamespace";

describe("RoleSelect", () => {
  let wrapper: VueWrapper<InstanceType<typeof RoleSelect>>;

  const mountWrapper = (modelValue = "administrator" as BasicRole) => {
    wrapper = mountComponent(RoleSelect, { props: { modelValue } });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => wrapper?.unmount());

  it("renders the role select component", () => {
    expect(wrapper.find('[data-test="role-select"]').exists()).toBe(true);
  });

  it("displays the initial model value", () => {
    const selectComponent = wrapper.findComponent({ name: "VSelect" });
    expect(selectComponent.props("modelValue")).toBe("administrator");
  });

  it("has all three role options available", () => {
    const select = wrapper.findComponent({ name: "VSelect" });
    const items = select.props("items");

    expect(items).toHaveLength(3);
    expect(items.map((item: { value: string }) => item.value)).toEqual([
      "administrator",
      "operator",
      "observer",
    ]);
  });

  it.each([
    {
      role: "administrator",
      title: "Administrator",
      description: "Full access to the namespace",
    },
    {
      role: "operator",
      title: "Operator",
      description: "Can manage and operate devices",
    },
    {
      role: "observer",
      title: "Observer",
      description: "Can view device details",
    },
  ])("has correct properties for $role role", ({ role, title, description }) => {
    const select = wrapper.findComponent({ name: "VSelect" });
    const items = select.props("items");
    const roleItem = items.find((item: { value: string }) => item.value === role);

    expect(roleItem.title).toBe(title);
    expect(roleItem.value).toBe(role);
    expect(roleItem.description).toContain(description);
  });

  it("emits update:modelValue when role is changed", async () => {
    const select = wrapper.findComponent({ name: "VSelect" });
    await select.vm.$emit("update:modelValue", "operator");

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["operator"]);
  });

  it("updates modelValue when a different role is selected", () => {
    mountWrapper("observer");

    const selectComponent = wrapper.findComponent({ name: "VSelect" });
    expect(selectComponent.props("modelValue")).toBe("observer");
  });
});
