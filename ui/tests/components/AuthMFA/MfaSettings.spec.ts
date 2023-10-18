import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaSettings from "@/components/AuthMFA/MfaSettings.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaSettingsWrapper = VueWrapper<InstanceType<typeof MfaSettings>>;

describe("MfaSettings", () => {
  let wrapper: MfaSettingsWrapper;

  const vuetify = createVuetify();

  let mock: MockAdapter;
  let mockNamespace: MockAdapter;

  // Define mock data for testing
  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];
  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };
  const authData = {
    status: "",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: false,
  };
  const mfaGenerate = {
    secret: "secret-mfa",
    link: "link-mfa",
    codes: [
      "HW2wlxV40B",
      "2xsmMUHHHb",
      "DTQgVsaVac",
      "KXPBoXvuWD",
      "QQYTPfotBi",
      "XWiKBEPyb4",
    ],
  };

  beforeEach(() => {
    // Use fake timers for testing
    vi.useFakeTimers();

    // Create a mock adapter for the mfaApi and namespacesApi instances
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mock = new MockAdapter(mfaApi.getAxios());

    // Mock API responses
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mock.onGet("http://localhost:3000/api/mfa/generate").reply(200, mfaGenerate);

    // Set initial state in the Vuex store
    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    // Mount the component with necessary dependencies
    wrapper = mount(MfaSettings, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
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

  it("Data is defined", () => {
    // Test if the component's data is defined
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Dialog opens", async () => {
    // Test if the dialog opens when a button is clicked
    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    expect(document.querySelector('[data-test="dialog"]')).not.toBeNull();
  });

  it("Renders the components (first-window)", async () => {
    // Test if the components in the first window are rendered as expected
    expect(wrapper.find('[data-test="enable-dialog-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    // Add more expectations for the first window components as needed
  });

  it("Renders the components (second-window)", async () => {
    // Test if the components in the second window are rendered as expected
    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    await wrapper.vm.goToNextStep();
    // Add more expectations for the second window components as needed
  });

  it("Renders the components (third-window)", async () => {
    // Test if the components in the third window are rendered as expected
    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    await wrapper.vm.goToNextStep();
    await wrapper.vm.goToNextStep();
    // Add more expectations for the third window components as needed
  });

  it("Successful MFA setup", async () => {
    // Test a successful MFA setup scenario
    const responseData = {
      token: "token",
    };

    // Mock the API response for MFA enable
    mock.onPost("http://localhost:3000/api/mfa/enable").reply(200, responseData);

    // Spy on the Vuex store dispatch method
    const mfaSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    await wrapper.vm.goToNextStep();

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    // Check if the store dispatch was called with the expected parameters
    expect(mfaSpy).toHaveBeenCalledWith("auth/enableMfa", {
      token_mfa: "000000",
      secret: "secret-mfa",
      codes: [
        "HW2wlxV40B",
        "2xsmMUHHHb",
        "DTQgVsaVac",
        "KXPBoXvuWD",
        "QQYTPfotBi",
        "XWiKBEPyb4",
      ],
    });
  });

  it("Error MFA setup", async () => {
    // Test an error scenario for MFA setup
    expect(wrapper.findComponent('[data-test="error-alert"]').exists()).toBe(false);

    // Mock an error response for MFA enable
    mock.onPost("http://localhost:3000/api/mfa/enable").reply(500);

    // Spy on the Vuex store dispatch method
    const mfaSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="enable-dialog-btn"]').trigger("click");
    await flushPromises();
    await wrapper.vm.goToNextStep();

    await wrapper.findComponent('[data-test="verification-code"]').setValue("000000");
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();

    // Check if the store dispatch was called with the expected parameters
    expect(mfaSpy).toHaveBeenCalledWith("auth/enableMfa", {
      token_mfa: "000000",
      secret: "secret-mfa",
      codes: [
        "HW2wlxV40B",
        "2xsmMUHHHb",
        "DTQgVsaVac",
        "KXPBoXvuWD",
        "QQYTPfotBi",
        "XWiKBEPyb4",
      ],
    });

    // Check if the error alert is displayed
    expect(wrapper.findComponent('[data-test="error-alert"]').exists()).toBe(true);
  });
});
