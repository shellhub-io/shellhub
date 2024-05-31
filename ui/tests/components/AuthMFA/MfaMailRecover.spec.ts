import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaMailRecover from "@/components/AuthMFA/MfaMailRecover.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaMailRecoverWrapper = VueWrapper<InstanceType<typeof MfaMailRecover>>;

describe("Mfa Mail Recover ", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: MfaMailRecoverWrapper;

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
    recovery_email: "recover@mail.com",
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
    wrapper = mount(MfaMailRecover, {
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

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });
});
