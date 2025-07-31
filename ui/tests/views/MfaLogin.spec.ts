import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaLogin from "@/views/MfaLogin.vue";
import { mfaApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import useAuthStore from "@/store/modules/auth";

type MfaLoginWrapper = VueWrapper<InstanceType<typeof MfaLogin>>;

describe("MfaLogin", () => {
  let wrapper: MfaLoginWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  beforeEach(() => {
    wrapper = mount(MfaLogin, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verification-code"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verify-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="redirect-recover"]').exists()).toBe(true);
  });

  it("disables submit button when the form is invalid", async () => {
    await wrapper.findComponent('[data-test="verification-code"]').setValue("");

    expect(wrapper.find('[data-test="verify-btn"]').attributes().disabled).toBeDefined();
  });

  it("calls the mfa action when the login form is submitted", async () => {
    const responseData = {
      token: "token",
    };

    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/auth").reply(200, responseData);

    const mfaSpy = vi.spyOn(authStore, "validateMfa");

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith("000000");
    expect(wrapper.vm.showAlert).toBe(false);
  });

  it("calls the mfa action when the login form is submitted (error)", async () => {
    const responseData = {
      token: "token",
    };

    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/auth").reply(500, responseData);

    const mfaSpy = vi.spyOn(authStore, "validateMfa");

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith("000000");
    expect(wrapper.vm.showAlert).toBe(true);
  });
});
