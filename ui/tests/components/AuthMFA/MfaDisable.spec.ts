import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaDisable from "@/components/AuthMFA/MfaDisable.vue";
import { mfaApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type MfaDisableWrapper = VueWrapper<InstanceType<typeof MfaDisable>>;

describe("MfaDisable", () => {
  let wrapper: MfaDisableWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(mfaApi.getAxios());

    wrapper = mount(MfaDisable, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],

      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
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

  it("Renders the components", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.showDialog = true;
    await flushPromises();

    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="verification-code"]').exists()).toBe(true);
    expect(dialog.find('[data-test="verify-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="use-recovery-code-btn"]').exists()).toBe(true);
    wrapper.vm.el = 2;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recovery-code"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recover-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="send-email-btn"]').exists()).toBe(true);
    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication using TOPT Code", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.showDialog = true;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalledWith({ code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
  });

  it("Disable MFA Authentication using TOPT Code (Fail)", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.showDialog = true;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();
    expect(mfaSpy).toHaveBeenCalledWith({ code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication using Recovery Code", async () => {
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();
    expect(mfaSpy).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
  });

  it("Disable MFA Authentication using Recovery Code (Fail)", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "disableMfa");
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Send the disable codes on the users mail", async () => {
    localStorage.setItem("email", "test@test.com");
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "requestMfaReset");
    mock.onPost("http://localhost:3000/api/user/mfa/reset").reply(200);
    await wrapper.findComponent('[data-test="send-email-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalled();
  });
  it("Send the disable codes on the users mail", async () => {
    const dialog = new DOMWrapper(document.body);

    localStorage.setItem("email", "test@test.com");
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    const mfaSpy = vi.spyOn(authStore, "requestMfaReset");
    mock.onPost("http://localhost:3000/api/user/mfa/reset").reply(403);
    await wrapper.findComponent('[data-test="send-email-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalled();
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });
});
