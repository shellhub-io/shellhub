import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Devices from "@/views/Devices.vue";
import { namespacesApi, usersApi, devicesApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type DevicesWrapper = VueWrapper<InstanceType<typeof Devices>>;

describe("Devices View", () => {
  let wrapper: DevicesWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockDevices: MockAdapter;

  const devices = [
    {
      uid: "a582b47a42d",
      name: "39-5e-2a",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "fake-tenant-data",
      last_seen: "2020-05-20T18:58:53.276Z",
      online: false,
      namespace: "user",
      status: "accepted",
    },
    {
      uid: "a582b47a42e",
      name: "39-5e-2b",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "fake-tenant-data",
      last_seen: "2020-05-20T19:58:53.276Z",
      online: true,
      namespace: "user",
      status: "accepted",
    },
  ];

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

  const res = {
    data: [namespaceData],
    headers: {
      "x-total-count": 1,
    },
  };

  const stats = {
    data: {
      registered_devices: 2,
      online_devices: 1,
      active_sessions: 0,
      pending_devices: 24,
      rejected_devices: 1,
    },
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10").reply(
      200,
      { data: devices, headers: { "x-total-count": 2 } },
    );
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("stats/setStats", stats);
    store.commit("devices/setDevices", { data: devices, headers: { "x-total-count": 2 } });

    wrapper = mount(Devices, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
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

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="device-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-header-component-group"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(false);
  });

  it("Shows the no items message when there are no public keys", async () => {
    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10").reply(200, []);
    store.commit("devices/setDevices", { data: [], headers: { "x-total-count": 0 } });
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').text()).toContain("Looks like you don't have any Devices");
  });
});
