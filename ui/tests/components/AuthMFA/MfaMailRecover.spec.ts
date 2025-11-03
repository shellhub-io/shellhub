import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import MfaMailRecover from "@/components/AuthMFA/MfaMailRecover.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaMailRecoverWrapper = VueWrapper<InstanceType<typeof MfaMailRecover>>;

describe("Mfa Mail Recover ", () => {
  let wrapper: MfaMailRecoverWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(MfaMailRecover, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });
});
