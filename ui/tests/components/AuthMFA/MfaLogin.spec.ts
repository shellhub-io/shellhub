import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaLogin from "@/components/AuthMFA/MfaLogin.vue";
import { mfaApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";

type MfaLoginWrapper = VueWrapper<InstanceType<typeof MfaLogin>>;

describe("MfaLogin", () => {
  let wrapper: MfaLoginWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    // Use fake timers and set a token in local storage
    vi.useFakeTimers();
    localStorage.setItem("token", "token");

    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(mfaApi.getAxios());

    // Mount the MfaLogin component with necessary dependencies
    wrapper = mount(MfaLogin, {
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
    expect(wrapper.find('[data-test="verification-code"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verify-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="redirect-recover"]').exists()).toBe(true);
  });

  it("disables submit button when the form is invalid", async () => {
    // Test if the submit button is disabled when the form is invalid
    await wrapper.findComponent('[data-test="verification-code"]').setValue("");

    // Check if the submit button has the 'disabled' attribute
    expect(wrapper.find('[data-test="verify-btn"]').attributes().disabled).toBeDefined();
  });

  it("calls the mfa action when the login form is submitted", async () => {
    // Test the scenario where the MFA login form is successfully submitted
    const responseData = {
      token: "token",
    };

    // Mock the API response for MFA authentication
    mock.onPost("http://localhost:3000/api/mfa/auth").reply(200, responseData);

    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    // Assert that the MFA authentication action was dispatched, and showAlert is false
    expect(mfaSpy).toHaveBeenCalledWith("auth/validateMfa", { code: "000000" });
    expect(wrapper.vm.showAlert).toBe(false);
  });

  it("calls the mfa action when the login form is submitted (error)", async () => {
    // Test the scenario where the MFA login form submission results in an error
    const responseData = {
      token: "token",
    };

    // Mock an error response for MFA authentication
    mock.onPost("http://localhost:3000/api/mfa/auth").reply(500, responseData);

    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    // Assert that the MFA authentication action was dispatched, and showAlert is true
    expect(mfaSpy).toHaveBeenCalledWith("auth/validateMfa", { code: "000000" });
    expect(wrapper.vm.showAlert).toBe(true);
  });
});
