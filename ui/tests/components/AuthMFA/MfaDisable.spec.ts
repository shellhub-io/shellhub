import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaDisable from "@/components/AuthMFA/MfaDisable.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaDisableWrapper = VueWrapper<InstanceType<typeof MfaDisable>>;

describe("MfaDisable", () => {
  let wrapper: MfaDisableWrapper;

  const vuetify = createVuetify();

  let mock: MockAdapter;

  let mockNamespace: MockAdapter;

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
    mfa: true,
  };

  beforeEach(() => {
    vi.useFakeTimers();
    // Create a mock adapter for the mfaApi and namespacesApi instances
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mock = new MockAdapter(mfaApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    // Commit auth and namespace data to the Vuex store
    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    // Mount the MfaDisable component with necessary dependencies
    wrapper = mount(MfaDisable, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
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
    // Test if component data is defined
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Dialog opens", async () => {
    // Test if the dialog opens when the button is clicked
    await wrapper.findComponent('[data-test="disable-dialog-btn"]').trigger("click");
    // Check if the dialog element is not null
    expect(document.querySelector('[data-test="dialog"]')).not.toBeNull();
  });

  it("Renders the components", async () => {
    // Test if the component's expected elements are rendered
    expect(wrapper.find('[data-test="disable-dialog-btn"]').exists()).toBe(true);
    // Open the dialog
    await wrapper.findComponent('[data-test="disable-dialog-btn"]').trigger("click");
    expect(wrapper.findComponent('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="dialog-text"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="close-btn"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="disable-btn"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication", async () => {
    // Test the scenario where MFA authentication is successfully disabled
    await wrapper.findComponent('[data-test="disable-dialog-btn"]').trigger("click");

    // Mock the API response for MFA disable
    mock.onPost("http://localhost:3000/api/mfa/disable").reply(200);

    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="disable-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa");
  });
});
