import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaResetValidation from "@/views/MfaResetValidation.vue";
import { mfaApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type MfaResetValidationWrapper = VueWrapper<InstanceType<typeof MfaResetValidation>>;

describe("Validate Recovery Mail", () => {
  let wrapper: MfaResetValidationWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  beforeEach(async () => {
    await router.push("/reset-mfa?id=xxxxxx");

    wrapper = mount(MfaResetValidation, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
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

  it("Renders the components", async () => {
    await flushPromises();

    expect(wrapper.find('[data-test="verification-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verification-subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verification-success"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="save-mail-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });

  it("Success on resetting the MFA", async () => {
    const storeSpy = vi.spyOn(authStore, "resetMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/reset/xxxxxx").reply(200, {});
    await wrapper.findComponent('[data-test="email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("1234");
    await wrapper.findComponent('[data-test="save-mail-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({ id: "xxxxxx", recovery_email_code: "1234", main_email_code: "123" });
  });

  it("Fails on resetting the MFA", async () => {
    const storeSpy = vi.spyOn(authStore, "resetMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/reset/xxxxxx").reply(403);
    await wrapper.findComponent('[data-test="email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="save-mail-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({ id: "xxxxxx", recovery_email_code: "123", main_email_code: "123" });
    expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(true);
  });
});
