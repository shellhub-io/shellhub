import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import WelcomeFourthScreen from "@/components/Welcome/WelcomeFourthScreen.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type WelcomeFourthScreenWrapper = VueWrapper<InstanceType<typeof WelcomeFourthScreen>>;

describe("Welcome Fourth Screen", () => {
  let wrapper: WelcomeFourthScreenWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(WelcomeFourthScreen, {
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

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="welcome-fourth-succesfully"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-fourth-links"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-fourth-thanks"]').exists()).toBe(true);
  });
});
