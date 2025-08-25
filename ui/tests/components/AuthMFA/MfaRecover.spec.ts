import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaRecover from "@/components/AuthMFA/MfaRecover.vue";
import { mfaApi } from "@/api/http";
import { router } from "@/router";
import useAuthStore from "@/store/modules/auth";

type MfaRecoverWrapper = VueWrapper<InstanceType<typeof MfaRecover>>;

describe("RecoverMFA", () => {
  let wrapper: MfaRecoverWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  beforeEach(() => {
    wrapper = mount(MfaRecover, {
      global: {
        plugins: [vuetify, router],
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
    expect(wrapper.find('[data-test="recovery-code"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="recover-btn"]').exists()).toBe(true);
  });

  it("disables submit button when the form is invalid", async () => {
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("");

    expect(wrapper.find('[data-test="recover-btn"]').attributes().disabled).toBeDefined();
  });

  it("calls the mfa action when the recover form is submitted", async () => {
    const responseData = {
      token: "token",
    };

    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/recover").reply(200, responseData);

    const mfaSpy = vi.spyOn(authStore, "recoverMfa");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="recovery-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith("000000");
    expect(routerPushSpy).toHaveBeenCalled();
  });

  it("calls the mfa action when the recover form is submitted", async () => {
    const responseData = {
      token: "token",
    };

    mockMfaApi.onPost("http://localhost:3000/api/user/mfa/recover").reply(403, responseData);

    const mfaSpy = vi.spyOn(authStore, "recoverMfa");

    await wrapper.findComponent('[data-test="recovery-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();

    expect(mfaSpy).toHaveBeenCalledWith("000000");
    expect(wrapper.vm.showAlert).toBe(true);
  });
});
