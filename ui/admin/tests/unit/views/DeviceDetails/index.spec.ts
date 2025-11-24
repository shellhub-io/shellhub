import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import routes from "@admin/router";
import DeviceDetails from "@admin/views/DeviceDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceDetailsWrapper = VueWrapper<InstanceType<typeof DeviceDetails>>;

const deviceDetail = {
  uid: "a582b47a42e",
  name: "39-5e-2b",
  identity: {
    mac: "00:00:00:00:00:00",
  },
  info: {
    id: "linuxmint",
    pretty_name: "Linux Mint 19.3",
    version: "v1.2.3",
    arch: "x86_64",
    platform: "linux",
  },
  public_key: "----- PUBLIC KEY -----",
  tenant_id: "fake-tenant-data",
  last_seen: "2020-05-20T19:58:53.276Z",
  online: true,
  namespace: "user",
  status: "accepted",
  created_at: "2020-05-01T00:00:00.000Z",
  remote_addr: "127.0.0.1",
  status_updated_at: "2020-05-02T00:00:00.000Z",
  position: {
    longitude: 0,
    latitude: 0,
  },
  tags: [
    {
      tenant_id: "fake-tenant-data",
      name: "test-tag",
      created_at: "",
      updated_at: "",
    },
  ],
};

const mockRoute = {
  params: {
    id: deviceDetail.uid,
  },
};

describe("Device Details", () => {
  let wrapper: DeviceDetailsWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();
    devicesStore.fetchDeviceById = vi.fn().mockResolvedValue(deviceDetail);

    const vuetify = createVuetify();

    wrapper = mount(DeviceDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: { $route: mockRoute },
      },
    });

    await flushPromises();
  });

  it("Displays the device name in the card title", () => {
    expect(wrapper.find(".text-h6").text()).toBe(deviceDetail.name);
  });

  it("Shows online status icon with tooltip", () => {
    const icon = wrapper.findComponent('[data-test="online-icon"]');
    expect(icon.classes()).toContain("mdi-check-circle");
    expect(icon.classes()).toContain("text-success");
  });

  it("Shows status chip with capitalized text", () => {
    const statusChip = wrapper.find('[data-test="device-status-chip"]');
    expect(statusChip.text()).toBe(deviceDetail.status);
  });

  it("Displays device UID", () => {
    const uidField = wrapper.find('[data-test="device-uid-field"]');
    expect(uidField.text()).toContain("UID:");
    expect(uidField.text()).toContain(deviceDetail.uid);
  });

  it("Displays MAC address", () => {
    const macField = wrapper.find('[data-test="device-mac-field"]');
    expect(macField.text()).toContain("MAC:");
    expect(macField.find("code").text()).toBe(deviceDetail.identity.mac);
  });

  it("Displays operating system with icon", () => {
    const osField = wrapper.find('[data-test="device-pretty-name-field"]');
    expect(osField.text()).toContain("Operating System:");
    expect(osField.text()).toContain(deviceDetail.info.pretty_name);
  });

  it("Displays agent version", () => {
    const versionField = wrapper.find('[data-test="device-version-field"]');
    expect(versionField.text()).toContain("Agent Version:");
    expect(versionField.text()).toContain(deviceDetail.info.version);
  });

  it("Displays architecture", () => {
    const archField = wrapper.find('[data-test="device-architecture-field"]');
    expect(archField.text()).toContain("Architecture:");
    expect(archField.text()).toContain(deviceDetail.info.arch);
  });

  it("Displays platform", () => {
    const platformField = wrapper.find('[data-test="device-platform-field"]');
    expect(platformField.text()).toContain("Platform:");
    expect(platformField.text()).toContain(deviceDetail.info.platform);
  });

  it("Displays namespace", () => {
    const namespaceField = wrapper.find('[data-test="device-namespace-field"]');
    expect(namespaceField.text()).toContain("Namespace:");
    expect(namespaceField.text()).toContain(deviceDetail.namespace);
  });

  it("Displays tenant ID", () => {
    const tenantField = wrapper.find('[data-test="device-tenant-id-field"]');
    expect(tenantField.text()).toContain("Tenant ID:");
    expect(tenantField.text()).toContain(deviceDetail.tenant_id);
  });

  it("Displays remote address", () => {
    const remoteAddrField = wrapper.find('[data-test="device-remote-addr-field"]');
    expect(remoteAddrField.text()).toContain("Remote Address:");
    expect(remoteAddrField.text()).toContain(deviceDetail.remote_addr);
  });

  it("Displays created at date", () => {
    const createdAtField = wrapper.find('[data-test="device-created-at-field"]');
    expect(createdAtField.text()).toContain("Created At:");
  });

  it("Displays device tags", () => {
    const tagsField = wrapper.find('[data-test="device-tags-field"]');
    expect(tagsField.text()).toContain("Tags:");
    expect(tagsField.text()).toContain(deviceDetail.tags[0].name);
  });

  it("Displays last seen date", () => {
    const lastSeenField = wrapper.find('[data-test="device-last-seen-field"]');
    expect(lastSeenField.text()).toContain("Last Seen:");
  });

  it("Displays public key", () => {
    const publicKeyField = wrapper.find('[data-test="device-public-key-field"]');
    expect(publicKeyField.text()).toContain("Public Key:");
    expect(publicKeyField.text()).toContain(deviceDetail.public_key);
  });

  it("Shows error message when device data is empty", async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();
    devicesStore.fetchDeviceById = vi.fn().mockResolvedValue({});

    const vuetify = createVuetify();

    const errorWrapper = mount(DeviceDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });

    await flushPromises();

    expect(errorWrapper.text()).toContain("Something is wrong, try again!");
  });
});
