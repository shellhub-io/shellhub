import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import SettingPrivateKeys from "@/components/Setting/SettingPrivateKeys.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingPrivateKeysWrapper = VueWrapper<InstanceType<typeof SettingPrivateKeys>>;

describe("Setting Private Keys", () => {
  let wrapper: SettingPrivateKeysWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(SettingPrivateKeys, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-button"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="private-key-list"]').exists()).toBe(true);
  });
});
