import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import Settings from "@/views/Settings.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type SettingsWrapper = VueWrapper<InstanceType<typeof Settings>>;

describe("Settings View", () => {
  let wrapper: SettingsWrapper;

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
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);

    wrapper = mount(Settings, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="settings-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="Profile-tab"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="Private Keys-tab"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="Tags-tab"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="Billing-tab"]').exists()).toBe(false);
  });

  it("Renders tabs based on visibleItems", async () => {
    await nextTick();
    const tabs = wrapper.findAllComponents({ name: "v-tab" });

    expect(tabs.length).toBe(wrapper.vm.visibleItems.length);
    for (let index = 0; index < tabs.length; index++) {
      expect(tabs[index].text()).toBe(wrapper.vm.visibleItems[index].title);
      expect(tabs[index].props("to")).toBe(wrapper.vm.visibleItems[index].path);
    }
  });
});
