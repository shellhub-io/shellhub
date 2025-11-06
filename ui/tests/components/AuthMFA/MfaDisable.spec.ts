import { nextTick } from "vue";
import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaDisable from "@/components/AuthMFA/MfaDisable.vue";
import { mfaApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type MfaDisableWrapper = VueWrapper<InstanceType<typeof MfaDisable>>;

describe("MfaDisable", () => {
  let wrapper: MfaDisableWrapper;
  let dialog: DOMWrapper<Element>;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(MfaDisable, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: { modelValue: true },
    });
    wrapper.vm.showDialog = true;
    await nextTick();
    dialog = new DOMWrapper(document.body);
    await flushPromises();
  });

  it("Renders the component (Verification Code window)", () => {
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the component (Recovery Code window)", async () => {
    wrapper.vm.el = 2;
    await flushPromises();
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the component (Email Sent window)", async () => {
    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Disables MFA Authentication using TOTP Code", async () => {
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalledWith({ code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
  });

  it("Disables MFA Authentication using TOTP Code (Fail)", async () => {
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();
    expect(mfaSpy).toHaveBeenCalledWith({ code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Disables MFA Authentication using Recovery Code", async () => {
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();
    expect(mfaSpy).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
  });

  it("Disables MFA Authentication using Recovery Code (Fail)", async () => {
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Sends the disable codes on the users mail", async () => {
    localStorage.setItem("email", "test@test.com");
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "requestMfaReset");
    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/reset").reply(200);
    await dialog.find('[data-test="recover-email-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalled();
  });

  it("Handles error when sending recovery email fails", async () => {
    localStorage.setItem("email", "test@test.com");
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "requestMfaReset");
    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/reset").reply(403);
    await dialog.find('[data-test="recover-email-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalled();
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });
});
