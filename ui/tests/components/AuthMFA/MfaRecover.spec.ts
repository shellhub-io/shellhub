import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaRecover from "@/components/AuthMFA/MfaRecover.vue";
import { mfaApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";

type MfaRecoverWrapper = VueWrapper<InstanceType<typeof MfaRecover>>;

describe("RecoverMFA", () => {
  let wrapper: MfaRecoverWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(mfaApi.getAxios());

    // Mount the MfaRecover component with necessary dependencies
    wrapper = mount(MfaRecover, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
    });
  });

  afterEach(() => {
    // Restore real timers and reset mocks after each test
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
  });

  it("Is a Vue instance", () => {
    // Test if the wrapper represents a Vue instance
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    // Test if the component renders as expected
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    // Test if the component's template contains expected elements
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="recovery-code"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="recover-btn"]').exists()).toBe(true);
  });

  it("disables submit button when the form is invalid", async () => {
    // Test if the submit button is disabled when the form is invalid
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("");

    // Check if the submit button has the 'disabled' attribute
    expect(wrapper.find('[data-test="recover-btn"]').attributes().disabled).toBeDefined();
  });

  it("calls the mfa action when the recover form is submitted", async () => {
    // Test the scenario where the MFA recovery form is successfully submitted
    const responseData = {
      token: "token",
    };

    // Mock the API response for MFA recovery
    mock.onPost("http://localhost:3000/api/mfa/recovery").reply(200, responseData);

    // Spy on Vuex store dispatch and router push methods
    const mfaSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="recovery-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();

    // Assert that the MFA recovery action was dispatched and router push was called
    expect(mfaSpy).toHaveBeenCalledWith("auth/recoverLoginMfa", { code: "000000" });
    expect(routerPushSpy).toHaveBeenCalled();
  });

  it("calls the mfa action when the recover form is submitted", async () => {
    // Test the scenario where the MFA recovery form submission results in an error
    const responseData = {
      token: "token",
    };

    // Mock an error response for MFA recovery
    mock.onPost("http://localhost:3000/api/mfa/recovery").reply(500, responseData);

    // Spy on Vuex store dispatch method
    const mfaSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="recovery-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();

    // Assert that the MFA recovery action was dispatched and the showAlert property is true
    expect(mfaSpy).toHaveBeenCalledWith("auth/recoverLoginMfa", { code: "000000" });
    expect(wrapper.vm.showAlert).toBe(true);
  });
});
