import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaSettings from "@/components/AuthMFA/MfaSettings.vue";
import { mfaApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type MfaSettingsWrapper = VueWrapper<InstanceType<typeof MfaSettings>>;

describe("MfaSettings", () => {
  let wrapper: MfaSettingsWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  const mfaGenerateData = {
    secret: "secret-mfa",
    link: "link-mfa",
    recovery_codes: [
      "HW2wlxV40B",
      "2xsmMUHHHb",
      "DTQgVsaVac",
      "KXPBoXvuWD",
      "QQYTPfotBi",
      "XWiKBEPyb4",
    ],
  };

  beforeEach(() => {
    mockMfaApi.onGet("http://localhost:3000/api/user/mfa/generate").reply(200, mfaGenerateData);

    wrapper = mount(MfaSettings, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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

  it("Dialog opens", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    expect(document.querySelector('[data-test="dialog"]')).not.toBeNull();
  });

  it("Renders the components (second-window)", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    wrapper.vm.goToNextStep();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components (third-window)", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    wrapper.vm.goToNextStep();
    wrapper.vm.goToNextStep();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Successful MFA setup", async () => {
    wrapper.vm.goToNextStep(); // 2
    const responseData = {
      token: "token",
    };
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/enable").reply(200, responseData);

    const mfaSpy = vi.spyOn(authStore, "enableMfa");

    wrapper.vm.showDialog = true;
    await flushPromises();
    wrapper.vm.goToNextStep(); // 3
    await flushPromises();
    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith({
      code: "000000",
      secret: mfaGenerateData.secret,
      recovery_codes: mfaGenerateData.recovery_codes,
    });
  });

  it("Error MFA setup", async () => {
    expect(wrapper.findComponent('[data-test="error-alert"]').exists()).toBe(false);
    wrapper.vm.goToNextStep(); // 2

    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/enable").reply(403);

    const mfaSpy = vi.spyOn(authStore, "enableMfa");

    wrapper.vm.showDialog = true;
    await flushPromises();
    wrapper.vm.goToNextStep(); // 3

    await flushPromises();

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith({
      code: "000000",
      secret: mfaGenerateData.secret,
      recovery_codes: mfaGenerateData.recovery_codes,
    });

    expect(wrapper.findComponent('[data-test="error-alert"]').exists()).toBe(true);
  });
});
