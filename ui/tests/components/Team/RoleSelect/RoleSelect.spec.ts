import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";

describe("RoleSelect", () => {
  let wrapper: VueWrapper<InstanceType<typeof RoleSelect>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(RoleSelect, {
      global: {
        plugins: [vuetify],
      },
      props: {
        modelValue: "administrator" as BasicRole,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders the role select component", () => {
    expect(wrapper.find('[data-test="role-select"]').exists()).toBe(true);
  });

  it("displays the initial model value", async () => {
    const selectComponent = wrapper.findComponent({ name: "VSelect" });
    expect(selectComponent.props("modelValue")).toBe("administrator");
  });

  it("has correct values for each role", () => {
    const select = wrapper.findComponent({ name: "VSelect" });
    const items = select.props("items");

    expect(items[0]).toEqual({
      title: "Administrator",
      value: "administrator",
      description: expect.stringContaining("Full access to the namespace"),
    });

    expect(items[1]).toEqual({
      title: "Operator",
      value: "operator",
      description: expect.stringContaining("Can manage and operate devices"),
    });

    expect(items[2]).toEqual({
      title: "Observer",
      value: "observer",
      description: expect.stringContaining("Can view device details"),
    });
  });
});
