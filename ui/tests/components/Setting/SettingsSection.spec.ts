import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import SettingsSection from "@/components/Setting/SettingsSection.vue";

describe("SettingsSection", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsSection>>;
  const vuetify = createVuetify();

  it("renders the slot content and forwards list attributes", () => {
    wrapper = mount(SettingsSection, {
      attrs: {
        "data-test": "settings-section-list",
      },
      slots: {
        default: '<div data-test="settings-section-item">Item</div>',
      },
      global: {
        plugins: [vuetify],
      },
    });

    expect(wrapper.find('[data-test="settings-section-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="settings-section-item"]').exists()).toBe(true);
  });
});
