import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { VApp } from "vuetify/components";
import AppBarContent from "@/components/AppBar/AppBarContent.vue";

const createWrapper = (props: Record<string, unknown> = {}) => {
  const vuetify = createVuetify();

  return mount({
    components: { AppBarContent, VApp },
    props,
    template: `
      <v-app>
        <AppBarContent v-bind="$props">
          <template #left>
            <div data-test="left-slot">Left Slot</div>
          </template>
          <template #right>
            <div data-test="right-slot">Right Slot</div>
          </template>
        </AppBarContent>
      </v-app>
    `,
  }, {
    props,
    global: {
      plugins: [vuetify],
    },
  });
};

describe("AppBarContent Component", () => {
  it("Renders slots and controls", () => {
    const wrapper = createWrapper({ showMenuToggle: true, showSupport: true });

    expect(wrapper.find('[data-test="app-bar"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="left-slot"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="right-slot"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="menu-toggle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="support-btn"]').exists()).toBe(true);
  });

  it("Emits toggle-menu and support-click", async () => {
    const wrapper = createWrapper({ showMenuToggle: true, showSupport: true });

    const appBar = wrapper.findComponent(AppBarContent);
    await wrapper.find('[data-test="menu-toggle"]').trigger("click");
    await wrapper.find('[data-test="support-btn"]').trigger("click");

    expect(appBar.emitted("toggle-menu")).toHaveLength(1);
    expect(appBar.emitted("support-click")).toHaveLength(1);
  });

  it("Hides controls when flags are false", () => {
    const wrapper = createWrapper({ showMenuToggle: false, showSupport: false });

    expect(wrapper.find('[data-test="menu-toggle"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="support-btn"]').exists()).toBe(false);
  });
});
