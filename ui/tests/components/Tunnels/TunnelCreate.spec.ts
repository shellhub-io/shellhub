import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TunnelCreate from "@/components/Tunnels/TunnelCreate.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi, tunnelApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type TunnelCreateWrapper = VueWrapper<InstanceType<typeof TunnelCreate>>;

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
    tags: ["test"],
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
    tags: ["test2"],
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
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const tunnelResponse = {
  address: "9a8df9321368d567cfac8679cec7848c",
  namespace: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  device: "13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a",
  host: "127.0.0.1",
  port: 8080,
};

describe("Tunnel Create", async () => {
  let wrapper: TunnelCreateWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;
  let mockTunnels: MockAdapter;

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockTunnels = new MockAdapter(tunnelApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("devices/setDeviceChooserStatus", true);

    wrapper = mount(TunnelCreate, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: "fake-uid",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component table", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="create-icon"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find('[data-test="tunnel-create-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tunnel-create-alert"]').exists()).toBe(false);
    expect(dialog.find('[data-test="tunnel-create-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="timeout-combobox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="address-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(false);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-tunnel-btn"]').exists()).toBe(true);
  });

  it("Successfully added tunnel", async () => {
    mockTunnels.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue(8080);
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(-1);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tunnels/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      ttl: -1,
      port: 8080,
    });
  });

  it("Successfully added tunnel (custom expiration)", async () => {
    mockTunnels.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");

    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue(8080);
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue("custom");
    await wrapper.findComponent('[data-test="custom-timeout"]').setValue(6000);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tunnels/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      ttl: 6000,
      port: 8080,
    });
  });

  it("Failed to add tunnel", async () => {
    mockTunnels.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(403);

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("bad-address");
    await wrapper.findComponent('[data-test="port-text"]').setValue("bad-port");
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(-1);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper
      .findComponent('[data-test="tunnel-create-alert"]').text()).toBe("This device has reached the maximum allowed number of tunnels");
  });
});
