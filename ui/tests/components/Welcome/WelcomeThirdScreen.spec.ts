import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import WelcomeThirdScreen from "@/components/Welcome/WelcomeThirdScreen.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type WelcomeThirdScreenWrapper = VueWrapper<InstanceType<typeof WelcomeThirdScreen>>;

describe("Welcome Third Screen", () => {
  let wrapper: WelcomeThirdScreenWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevices : MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    devices: 2,
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
  };

  const stats = {
    registered_devices: 2,
    online_devices: 1,
    active_sessions: 0,
    pending_devices: 24,
    rejected_devices: 1,
  };

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
      status: "pending",
    },
  ];

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);
    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending").reply(200, devices);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(WelcomeThirdScreen, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
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

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="welcome-third-screen-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-hostname"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-os"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-pretty-name-field"]').exists()).toBe(true);
  });
});
