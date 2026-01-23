import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import SettingsRow from "@/components/Setting/SettingsRow.vue";

describe("SettingsRow", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsRow>>;
  const vuetify = createVuetify();

  const mountComponent = () => {
    wrapper = mount(SettingsRow, {
      props: {
        icon: "mdi-test-icon",
        iconTestId: "row-icon",
        title: "Row Title",
        titleTestId: "row-title",
        subtitle: "Row subtitle",
        subtitleTestId: "row-subtitle",
      },
      slots: {
        default: '<div data-test="row-field">Field</div>',
      },
      global: {
        plugins: [vuetify],
      },
    });
  };

  it("renders icon, title, subtitle, and field slot", () => {
    mountComponent();

    const icon = wrapper.find('[data-test="row-icon"]');
    expect(icon.exists()).toBe(true);
    expect(icon.classes().join(" ")).toContain("mdi-test-icon");
    expect(wrapper.find('[data-test="row-title"]').text()).toBe("Row Title");
    expect(wrapper.find('[data-test="row-subtitle"]').text()).toBe("Row subtitle");
    expect(wrapper.find('[data-test="row-field"]').exists()).toBe(true);
  });
});
