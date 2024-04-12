import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ValidationAccount from "@/views/ValidationAccount.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type ValidationAccountWrapper = VueWrapper<InstanceType<typeof ValidationAccount>>;

describe("Validation Account", () => {
  let wrapper: ValidationAccountWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

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
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  const res = {
    data: [namespaceData],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(async () => {
    await router.push("/validation-account?email=test@test.com&token=test-token");
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/user/validation_account?email=test%40test.com&token=test-token").reply(200);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);

    wrapper = mount(ValidationAccount, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
    mockUser.reset();
  });

  it("Is a Vue instance", async () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", async () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="verification-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="processing-cardText"]').exists()).toBe(true);
  });

  it("Renders success message", async () => {
    await flushPromises();
    expect(wrapper.find('[data-test="success-cardText"]').exists()).toBe(true);
  });

  it("Redirects to login page when login link is clicked", async () => {
    await flushPromises();
    await wrapper.find('[data-test="login-btn"]').trigger("click");
    expect(router.currentRoute.value.path).toBe("/validation-account");
  });
});
